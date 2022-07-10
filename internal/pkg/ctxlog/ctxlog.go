package ctxlog

import (
	"context"
	"fmt"
	"os"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// context key for logger extraction, it is recommended not to use any build-in
// type for context keys to avoid collision
type loggerKeyType struct{}

const (
	// defaultEnvName is the default loglevel if initialization did not specify
	defaultEnvName = "LOG_LEVEL"

	// DefaultLevel is the the default environment name
	DefaultLevel = zapcore.InfoLevel
)

var (
	_loggerKey = loggerKeyType{}

	// global logger
	_logger *zap.Logger

	// zap log level
	_level zap.AtomicLevel
)

// InitializeLoggerFromEnv initializes the global
func InitializeLoggerFromEnv(envName ...string) *zap.Logger {
	var level string

	// check for env name override
	if len(envName) == 0 {
		level = os.Getenv(defaultEnvName)
	} else {
		level = os.Getenv(envName[0])
	}

	switch level {
	case "debug", "DEBUG", "Debug":
		return InitializeLogger(zap.DebugLevel)
	case "info", "INFO", "Info":
		return InitializeLogger(zap.InfoLevel)
	case "warn", "WARN", "Warn":
		return InitializeLogger(zap.WarnLevel)
	case "error", "ERROR", "Error":
		return InitializeLogger(zap.ErrorLevel)
	default:
		return InitializeLogger()
	}
}

// InitializeLogger initializes the global logger
func InitializeLogger(initLevel ...zapcore.Level) *zap.Logger {
	// first check if it's already initialized, if logger is already present,
	// don't re-initialize, only adjust level if it's given
	if _logger != nil {
		if len(initLevel) > 0 {
			_level.SetLevel(initLevel[0])
		}
		return _logger
	}

	// change automatic timestamp field
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.DisableStacktrace = true
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerConfig.EncoderConfig.MessageKey = "message"

	_level = zap.NewAtomicLevel()
	loggerConfig.Level = _level

	// if there is an initial level being pased in
	if len(initLevel) > 0 {
		_level.SetLevel(initLevel[0])
	} else {
		_level.SetLevel(DefaultLevel)
	}

	var err error
	_logger, err = loggerConfig.Build()

	if err != nil {
		fmt.Println("Unable to initialize logger")
		panic(err)
	}

	_logger.Info("logger initialized")

	return bundleLogger(_logger, false)
}

// Logger retrieves the logger from context, and attaches the specified call
// stack to the logger. Generally callstack should be set to the
// function/method name you're calling from to make log trace easier.
// It is preferred over the LoggerRaw
func Logger(ctx context.Context, callstack string, fields ...zap.Field) *zap.Logger {
	return LoggerRaw(ctx).Named(callstack).With(fields...)
}

// WrapCtx create new context with embedded logger with callstack
// tracing. This can be used to instantiate a new context within a
// function/methods with call tracing
func WrapCtx(ctx context.Context, callstack string, fields ...zap.Field) context.Context {
	return context.WithValue(
		ctx,
		_loggerKey,
		Logger(ctx, callstack).With(fields...),
	)
}

// WrapCtxLogger combines callstack setting and logger retrival.
// Normally callstack should be the function name, but can be others if needed.
// This is useful to be placed at the beginning of your function so that you
// would have a local context with callstack trace and a logger that is set
// with structured field and callstack appended
func WrapCtxLogger(
	ctx context.Context,
	callstack string,
	fields ...zap.Field,
) (newctx context.Context, logger *zap.Logger) {
	newctx = WrapCtx(ctx, callstack, fields...)
	logger = LoggerRaw(newctx)
	return
}

// LoggerRaw extract the embedded contextual logger, if not, check to see
// if gRPC context logger is present, if not return the global logger. You
// should prefer Logger because it expands callstack and this
// function does not.
func LoggerRaw(ctx context.Context) (logger *zap.Logger) {

	// if the top-level context is nil, ctx.Value(...) causes a panic, ideally
	// this shouldn't happen, but I think somewhere in our toolchain a nil
	// context is being used for the top level context, we have to handle
	// this case and treat this as no existing loggers found
	defer func() {
		if r := recover(); r != nil {
			logger = manifestLogger(ctx)
		}
	}()

	// ensure we have a global logger, under normal circumstances this shouldn't
	// be necessary, but it's added for protection purpose
	if _logger == nil {
		InitializeLoggerFromEnv()
	}

	// nil context yield a global logger with new request_id
	if ctx == nil {
		return bundleLogger(_logger, true)
	}

	// existing context logger is found
	if ctxLogger, ok := ctx.Value(_loggerKey).(*zap.Logger); ok {
		return ctxLogger
	}

	// no existing ctx logger found, we will manifest a new one
	return manifestLogger(ctx)
}

// manifestLogger attempts to retrieve ctxzap logger and attaches a new request
// ID, if it fails or panics, it will return the root logger with a new request
// ID
func manifestLogger(_ context.Context) (logger *zap.Logger) {

	// same as function above, top-level nil context panic has to be
	// handled since ctxzap.Extract() itself does not handle the panic
	defer func() {
		if r := recover(); r != nil {
			logger = bundleLogger(_logger, true)
		}
	}()

	// see if we can extract GRPC ctx logger
	// ctxZap will return a null logger if it's not found, that's not what we
	// want because a null logger will not output any logs. By checking against
	// a known null logger (extracted from background context), we'll be able
	// to determine that
	//if ctxZap := ctxzap.Extract(ctx); ctxZap != ctxzap.Extract(context.Background()) {
	//	// bundle standard fields
	//	return bundleLogger(ctxZap, true)
	//}

	// context doesn't have what we need, start a new logger
	return bundleLogger(_logger, true)
}

// bundleLogger takes the logger and returns with one with pre-populated fields,
// set newRID to true will cause a new request_id generated and inserted
func bundleLogger(logger *zap.Logger, newRID bool) *zap.Logger {
	bundled := logger

	if newRID {
		return bundled.With(zap.String("request_id", uuid.NewV4().String()))
	}

	return bundled
}

// WrapCtxRaw create new context with embedded logger but no callstack trace.
// Prefer WrapCtx over this as it will give a callstack trace
// as well.
func WrapCtxRaw(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(
		ctx,
		_loggerKey,
		LoggerRaw(ctx).With(fields...),
	)
}

// SetLevelDebug sets log level to debug
func SetLevelDebug() {
	// ensure we have a global logger, under normal circumstances this shouldn't
	// be necessary, but it's added for protection purpose
	if _logger == nil {
		InitializeLoggerFromEnv()
	}

	_level.SetLevel(zap.DebugLevel)
}

// SetLevelInfo sets log level to info
func SetLevelInfo() {
	// ensure we have a global logger, under normal circumstances this shouldn't
	// be necessary, but it's added for protection purpose
	if _logger == nil {
		InitializeLoggerFromEnv()
	}

	_level.SetLevel(zap.InfoLevel)
}

// SetLevelWarn sets log level to warn
func SetLevelWarn() {
	// ensure we have a global logger, under normal circumstances this shouldn't
	// be necessary, but it's added for protection purpose
	if _logger == nil {
		InitializeLoggerFromEnv()
	}

	_level.SetLevel(zap.WarnLevel)
}

// SetLevelError sets log level to eror
func SetLevelError() {
	// ensure we have a global logger, under normal circumstances this shouldn't
	// be necessary, but it's added for protection purpose
	if _logger == nil {
		InitializeLoggerFromEnv()
	}

	_level.SetLevel(zap.ErrorLevel)
}

// GetRootLogger retrieves the global logger
func GetRootLogger() *zap.Logger {
	// ensure we have a global logger, under normal circumstances this shouldn't
	// be necessary, but it's added for protection purpose
	if _logger == nil {
		InitializeLoggerFromEnv()
	}

	return bundleLogger(_logger, false)
}

// GetZapLevel get zap level for grpc logger level option
func GetZapLevel() zapcore.Level {
	return _level.Level()
}
