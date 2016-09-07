package messages

import (
	"testing"
	"bytes"
	"time"
)

func TestClientIncomingHello(t *testing.T) {
	incoming := []byte("HELLO 1.0.0 mynonce")
	_, msg, err := LoadMessage("client_incoming", incoming)

	if err != nil {
		t.Error("Failed to parse ClientIncomingHello")
		return
	}

	cih, ok := msg.(*ClientIncomingHello)

	if !ok {
		t.Error("Failed to receive ClientIncomingHello")
		return
	}

	if cih.Version != "1.0.0" {
		t.Error("Failed to parse version")
		return
	}

	if cih.Nonce != "mynonce" {
		t.Error("Failed to parse nonce")
		return
	}

	outgoing := cih.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}

func TestClientIncomingOn(t *testing.T) {
	incoming := []byte("ON lock 123 mynonce")
	_, msg, err := LoadMessage("client_incoming", incoming)

	if err != nil {
		t.Error("Failed to parse ClientIncomingOn")
		return
	}

	cih, ok := msg.(*ClientIncomingOn)

	if !ok {
		t.Error("Failed to receive ClientIncomingOn")
		return
	}

	if cih.Lock != "lock" {
		t.Error("Failed to parse lock")
		return
	}

	if cih.Timeout != time.Millisecond * 123 {
		t.Error("Failed to parse timeout")
		return
	}

	if cih.Nonce != "mynonce" {
		t.Error("Failed to parse nonce")
		return
	}

	outgoing := cih.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}

func TestClientIncomingOff(t *testing.T) {
	incoming := []byte("OFF lock mynonce")
	_, msg, err := LoadMessage("client_incoming", incoming)

	if err != nil {
		t.Error("Failed to parse ClientIncomingOff")
		return
	}

	cih, ok := msg.(*ClientIncomingOff)

	if !ok {
		t.Error("Failed to receive ClientIncomingOff")
		return
	}

	if cih.Lock != "lock" {
		t.Error("Failed to parse lock")
		return
	}

	if cih.Nonce != "mynonce" {
		t.Error("Failed to parse nonce")
		return
	}

	outgoing := cih.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}
