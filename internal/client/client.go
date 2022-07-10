package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/bolom009/tcp-pow/internal/pkg/ctxlog"
	"go.uber.org/zap"

	"github.com/bolom009/tcp-pow/internal/pkg/config"
	"github.com/bolom009/tcp-pow/internal/pkg/pow"
	"github.com/bolom009/tcp-pow/internal/pkg/protocol"
)

type Client struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
	}
}

func (c *Client) Start(ctx context.Context, address string) error {
	ctx, logger := ctxlog.WrapCtxLogger(ctx, "Start", zap.String("address", address))

	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Error("FailedToDial", zap.Error(err))
		return err
	}

	logger.Info("ConnectedToAddress")

	defer conn.Close()

	for {
		message, err := c.handleConnection(ctx, conn, conn)
		if err != nil {
			logger.Error("FailedToHandleConnection", zap.Error(err))
			return err
		}

		logger.Info("Result", zap.String("quote", message))
		time.Sleep(5 * time.Second)
	}
}

// handleConnection - scenario for TCP-client
// 1. request challenge from server
// 2. compute hashcash to check POW
// 3. send hashcash solution back to server
// 4. get result quote from server
func (c *Client) handleConnection(ctx context.Context, readerConn io.Reader, writerConn io.Writer) (string, error) {
	logger := ctxlog.Logger(ctx, "handleConnection")

	reader := bufio.NewReader(readerConn)

	// 1. requesting challenge
	err := sendMsg(protocol.Message{Header: protocol.RequestChallenge}, writerConn)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	// reading and parsing response challenge-response
	msgStr, err := readMsg(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read msg: %w", err)
	}

	msg, err := protocol.ParseMessage(msgStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse msg: %w", err)
	}

	var hashcash pow.HashcashData
	err = json.Unmarshal([]byte(msg.Payload), &hashcash)
	if err != nil {
		return "", fmt.Errorf("failed to parse hashcash: %w", err)
	}

	logger.Info("GotChallengeResponse", zap.String("hashcash", hashcash.Stringify()))

	// 2. got challenge, compute hashcash
	hashcash, err = hashcash.ComputeHashcash(c.cfg.Hashcash.MaxIterations)
	if err != nil {
		return "", fmt.Errorf("failed to compute hashcash: %w", err)
	}

	logger.Info("ComputedHashcash", zap.String("hashcash", hashcash.Stringify()))

	byteData, err := json.Marshal(hashcash)
	if err != nil {
		return "", fmt.Errorf("failed to marshal hashcash: %w", err)
	}

	// 3. send computed challenge back to server
	err = sendMsg(protocol.Message{Header: protocol.RequestResource, Payload: string(byteData)}, writerConn)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	logger.Info("ChallengeSentToServer", zap.String("hashcash", hashcash.Stringify()))

	// 4. get result quote from server
	msgStr, err = readMsg(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read msg: %w", err)
	}

	msg, err = protocol.ParseMessage(msgStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse msg: %w", err)
	}

	return msg.Payload, nil
}

func readMsg(reader *bufio.Reader) (string, error) {
	return reader.ReadString('\n')
}

func sendMsg(msg protocol.Message, conn io.Writer) error {
	msgStr := fmt.Sprintf("%s\n", msg.Stringify())
	_, err := conn.Write([]byte(msgStr))
	return err
}
