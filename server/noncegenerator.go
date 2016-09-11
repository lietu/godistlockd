package server

import (
	"encoding/base64"
	"encoding/binary"
	"bytes"
)

type NonceGenerator struct {
	nonceChan chan uint64
}

func (ng *NonceGenerator) Uint64() uint64 {
	return <-ng.nonceChan
}

func (ng *NonceGenerator) String() string {
	nonce := ng.Uint64()
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffer, nonce)

	buffer = bytes.Trim(buffer, "\x00")

	return base64.StdEncoding.EncodeToString(buffer)
}

func (ng *NonceGenerator) nonceGenerator() {
	var i uint64
	for i = 1; ; i += 1 {
		ng.nonceChan <- i
	}
}

func NewNonceGenerator() *NonceGenerator {
	ng := NonceGenerator{}
	ng.nonceChan = make(chan uint64)

	go ng.nonceGenerator()

	return &ng
}
