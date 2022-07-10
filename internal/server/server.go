package server

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/bolom009/tcp-pow/internal/pkg/config"
	"github.com/bolom009/tcp-pow/internal/pkg/ctxlog"
	"github.com/bolom009/tcp-pow/internal/pkg/db"
	"github.com/bolom009/tcp-pow/internal/pkg/pow"
	"github.com/bolom009/tcp-pow/internal/pkg/protocol"
	"go.uber.org/zap"
)

var quotes = []string{
	"Sniper here! Time for target practice.",
	"Your mind is mine! Not like you were using it anyway.",
	"There are some things you weren't meant to see. Such as tomorrow.",
	"Roshan!!! I come to reclaim what you stole!!",
	"One Puck, two Puck, three Puck more!", // the best :)
}

var (
	ErrQuit              = errors.New("client requests to close connection")
	ErrInvalidResource   = errors.New("invalid hashcash resource")
	ErrNotFoundChallenge = errors.New("challenge not sent")
	ErrExpiredChallenge  = errors.New("challenge expired")
	ErrInvalidHashcash   = errors.New("invalid hashcash")
)

type Server struct {
	cache db.Cache
	cfg   *config.Config
}

func New(cfg *config.Config, cache db.Cache) *Server {
	return &Server{
		cfg:   cfg,
		cache: cache,
	}
}

func (s *Server) Start(ctx context.Context, address string) error {
	ctx, logger := ctxlog.WrapCtxLogger(ctx, "Start")

	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error("FailedToRunListen", zap.String("address", address), zap.Error(err))
		return err
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("FailedToAcceptClient", zap.Error(err))
			return fmt.Errorf("failed accept connection: %w", err)
		}

		go s.handleConnection(ctx, conn)
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	ctx, logger := ctxlog.WrapCtxLogger(ctx, "handleConnection", zap.String("remoteAddr", conn.RemoteAddr().String()))

	logger.Info("ConnectedNewClient")

	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		req, err := reader.ReadString('\n')
		if err != nil {
			logger.Error("FailedToReadMessage", zap.Error(err))
			return
		}

		msg, err := s.processRequest(ctx, req, conn.RemoteAddr().String())
		switch err {
		case ErrQuit:
			return
		case nil:
			break
		default:
			logger.Error("FailedToProcessRequest", zap.Error(err))
			return
		}

		if msg == nil {
			logger.Warn("MessageEmpty", zap.Error(err))
			return
		}

		if err := sendMsg(*msg, conn); err != nil {
			logger.Error("FailedToSendMessage", zap.Error(err))
		}
	}
}

func (s *Server) processRequest(ctx context.Context, msgStr string, clientInfo string) (*protocol.Message, error) {
	ctx, logger := ctxlog.WrapCtxLogger(ctx, "processRequest")

	msg, err := protocol.ParseMessage(msgStr)
	if err != nil {
		logger.Error("FailedToParseMessage", zap.String("message", msgStr), zap.Error(err))
		return nil, err
	}

	switch msg.Header {
	case protocol.Quit:
		return nil, ErrQuit
	case protocol.RequestChallenge:
		logger.Info("ClientSentRequestChallenge")

		// mechanism for check timing of challenge
		uid := uuid.NewV4()
		id := uid.String()
		err := s.cache.Add(ctx, id, time.Duration(s.cfg.Hashcash.Duration))
		if err != nil {
			logger.Error("FailedToAddToCacheDuration", zap.Error(err))
			return nil, fmt.Errorf("failed to add rand to cache: %w", err)
		}

		hashcash := pow.NewHashcash(s.cfg.Hashcash.ZeroCount, clientInfo, id)
		b, err := json.Marshal(hashcash)
		if err != nil {
			logger.Error("FailedToMarshalHashcash", zap.Error(err))
			return nil, fmt.Errorf("failed to marshal hashcash: %v", err)
		}

		return &protocol.Message{
			Header:  protocol.ResponseChallenge,
			Payload: string(b),
		}, nil
	case protocol.RequestResource:
		logger.Info("ClientSentRequestResource", zap.String("payload", msg.Payload))

		var hashcash pow.HashcashData
		if err := json.Unmarshal([]byte(msg.Payload), &hashcash); err != nil {
			logger.Error("FailedToUnmarshalPayload", zap.Error(err))
			return nil, fmt.Errorf("failed to unmarshal hashcash: %w", err)
		}

		if hashcash.Resource != clientInfo {
			logger.Warn("ClientSentWrongParamResource", zap.String("clientInfo", clientInfo),
				zap.String("resource", hashcash.Resource))
			return nil, ErrInvalidResource
		}

		randValueBytes, err := base64.StdEncoding.DecodeString(hashcash.Rand)
		if err != nil {
			logger.Error("FailedToStdDecode", zap.Error(err))
			return nil, fmt.Errorf("failed to decode rand: %w", err)
		}

		uid := string(randValueBytes)

		exists, err := s.cache.Exist(ctx, uid)
		if err != nil {
			logger.Error("FailedToGetCacheValue", zap.Error(err))
			return nil, fmt.Errorf("failed to get rand from cache: %w", err)
		}

		if !exists {
			return nil, ErrNotFoundChallenge
		}

		if time.Now().Unix()-hashcash.Date > s.cfg.Hashcash.Duration {
			return nil, ErrExpiredChallenge
		}

		maxIter := hashcash.Counter
		if maxIter <= 0 {
			logger.Warn("ClientSentWrongParamCounter", zap.Int("counter", maxIter))
			return nil, ErrInvalidHashcash
		}

		_, err = hashcash.ComputeHashcash(maxIter)
		if err != nil {
			logger.Error("FailedToComputeHashcash", zap.Int("counter", maxIter), zap.Error(err))
			return nil, ErrInvalidHashcash
		}

		randQuote := rand.Intn(len(quotes))

		logger.Info("ClientSuccessfullyComputedHashcash", zap.String("payload", msg.Payload))
		logger.Info("SendClientQuote", zap.String("quote", quotes[randQuote]))

		if err := s.cache.Delete(ctx, uid); err != nil {
			logger.Warn("FailedToDeleteCache", zap.String("key", uid), zap.Error(err))
		}

		return &protocol.Message{
			Header:  protocol.ResponseResource,
			Payload: quotes[randQuote],
		}, nil
	default:
		return nil, fmt.Errorf("unknown header")
	}
}

// sendMsg - send protocol message to connection
func sendMsg(msg protocol.Message, conn net.Conn) error {
	msgStr := fmt.Sprintf("%s\n", msg.Stringify())
	_, err := conn.Write([]byte(msgStr))
	return err
}
