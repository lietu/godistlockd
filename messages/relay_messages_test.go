package messages

import (
	"testing"
	"bytes"
	"time"
)

func TestRelayIncomingHello(t *testing.T) {
	incoming := []byte("HELLO server-1 1.0.0 mynonce")
	_, genmsg, err := LoadMessage("relay", incoming)

	if err != nil {
		t.Error("Failed to parse RelayIncomingHello")
		return
	}

	msg, ok := genmsg.(*RelayIncomingHello)

	if !ok {
		t.Error("Failed to receive RelayIncomingHello")
		return
	}

	if msg.Id != "server-1" {
		t.Error("Failed to parse id")
	}

	if msg.Version != "1.0.0" {
		t.Error("Failed to parse version")
	}

	if msg.Nonce != "mynonce" {
		t.Error("Failed to parse nonce")
	}

	outgoing := msg.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}

func TestRelayIncomingProp(t *testing.T) {
	incoming := []byte("PROP lock-1 nonce-1")
	_, genmsg, err := LoadMessage("relay", incoming)

	if err != nil {
		t.Error("Failed to parse RelayIncomingProp")
		return
	}

	msg, ok := genmsg.(*RelayIncomingProp)

	if !ok {
		t.Error("Failed to receive RelayIncomingProp")
		return
	}

	if msg.Lock != "lock-1" {
		t.Error("Failed to parse lock")
	}

	if msg.Nonce != "nonce-1" {
		t.Error("Failed to parse nonce")
	}

	outgoing := msg.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}

func TestRelayIncomingSched(t *testing.T) {
	incoming := []byte("SCHED lock-1 nonce-1")
	_, genmsg, err := LoadMessage("relay", incoming)

	if err != nil {
		t.Error("Failed to parse RelayIncomingSched")
		return
	}

	msg, ok := genmsg.(*RelayIncomingSched)

	if !ok {
		t.Error("Failed to receive RelayIncomingSched")
		return
	}

	if msg.Lock != "lock-1" {
		t.Error("Failed to parse lock")
	}

	if msg.Nonce != "nonce-1" {
		t.Error("Failed to parse nonce")
	}

	outgoing := msg.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}


func TestRelayIncomingComm(t *testing.T) {
	incoming := []byte("COMM lock-1 123 nonce-1")
	_, genmsg, err := LoadMessage("relay", incoming)

	if err != nil {
		t.Error("Failed to parse RelayIncomingComm")
		return
	}

	msg, ok := genmsg.(*RelayIncomingComm)

	if !ok {
		t.Error("Failed to receive RelayIncomingComm")
		return
	}

	if msg.Lock != "lock-1" {
		t.Error("Failed to parse lock")
	}

	if msg.Timeout != time.Millisecond * 123 {
		t.Error("Failed to parse timeout")
	}

	if msg.Nonce != "nonce-1" {
		t.Error("Failed to parse nonce")
	}

	outgoing := msg.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}


func TestRelayIncomingOff(t *testing.T) {
	incoming := []byte("OFF lock-1 nonce-1")
	_, genmsg, err := LoadMessage("relay", incoming)

	if err != nil {
		t.Error("Failed to parse RelayIncomingOff")
		return
	}

	msg, ok := genmsg.(*RelayIncomingOff)

	if !ok {
		t.Error("Failed to receive RelayIncomingOff")
		return
	}

	if msg.Lock != "lock-1" {
		t.Error("Failed to parse lock")
	}

	if msg.Nonce != "nonce-1" {
		t.Error("Failed to parse nonce")
	}

	outgoing := msg.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}
