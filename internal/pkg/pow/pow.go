package pow

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

const (
	DefaultVersion = 1
	DefaultCounter = 0

	zeroByte = 48
)

var (
	ErrMaxIterations = errors.New("max iterations exceeded")
)

type HashcashData struct {
	Version    int
	ZerosCount int
	Date       int64
	Resource   string
	Rand       string
	Counter    int
}

func NewHashcash(zeroCount int, clientInfo string, uid string) *HashcashData {
	return &HashcashData{
		Version:    DefaultVersion,
		ZerosCount: zeroCount,
		Date:       time.Now().Unix(),
		Resource:   clientInfo,
		Rand:       base64.StdEncoding.EncodeToString([]byte(uid)),
		Counter:    DefaultCounter,
	}
}

func (h HashcashData) Stringify() string {
	return fmt.Sprintf("%d:%d:%d:%s::%s:%d", h.Version, h.ZerosCount, h.Date, h.Resource, h.Rand, h.Counter)
}

func sha1Hash(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

func IsHashCorrect(hash string, zerosCount int) bool {
	if zerosCount > len(hash) {
		return false
	}

	for _, ch := range hash[:zerosCount] {
		if ch != zeroByte {
			return false
		}
	}

	return true
}

// ComputeHashcash is a cryptographic hash-based proof-of-work algorithm
// that requires a selectable amount of work to compute, but the proof can be verified efficiently
func (h *HashcashData) ComputeHashcash(maxIterations int) (HashcashData, error) {
	for h.Counter <= maxIterations || maxIterations <= 0 {
		if h.Verify() {
			return *h, nil
		}

		h.Counter++
	}
	return *h, ErrMaxIterations
}

func (h HashcashData) Verify() bool {
	header := h.Stringify()
	hash := sha1Hash(header)

	return IsHashCorrect(hash, h.ZerosCount)
}
