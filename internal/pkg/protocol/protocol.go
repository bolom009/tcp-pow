package protocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// type of tcp messages for communicate client-server
const (
	Quit = iota + 1
	RequestChallenge
	ResponseChallenge
	RequestResource
	ResponseResource
)

const splitter = "|"

var (
	ErrInvalidHeader  = errors.New("cannot parse header")
	ErrInvalidMessage = errors.New("invalid message")
)

type Message struct {
	Header  int
	Payload string
}

func (m *Message) Stringify() string {
	return fmt.Sprintf("%d%s%s", m.Header, splitter, m.Payload)
}

func ParseMessage(str string) (*Message, error) {
	str = strings.TrimSpace(str)

	parts := strings.Split(str, splitter)
	if len(parts) < 1 || len(parts) > 2 {
		return nil, ErrInvalidMessage
	}

	msgType, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, ErrInvalidHeader
	}

	msg := Message{Header: msgType}
	if len(parts) == 2 {
		msg.Payload = parts[1]
	}

	return &msg, nil
}
