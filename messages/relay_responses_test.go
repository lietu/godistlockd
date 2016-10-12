package messages

import (
	"testing"
	"bytes"
)

func TestRelayHowdy(t *testing.T) {
	incoming := []byte("HOWDY nonce server-1 1.0.0")
	_, genmsg, err := LoadMessage("relay", incoming)

	if err != nil {
		t.Error("Failed to create RelayHowdy")
		return
	}

	msg, ok := genmsg.(*RelayHowdy)

	if !ok {
		t.Error("Failed to receive RelayHowdy")
		return
	}

	if msg.Nonce != "nonce" {
		t.Error("Failed to parse nonce")
	}

	if msg.Id != "server-1" {
		t.Error("Failed to parse server")
	}

	if msg.Version != "1.0.0" {
		t.Error("Failed to parse version")
	}

	outgoing := genmsg.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}

func TestRelayStat(t *testing.T) {
	incoming := []byte("STAT nonce 0")
	_, genmsg, err := LoadMessage("relay", incoming)

	if err != nil {
		t.Error("Failed to create RelayStat")
		return
	}

	msg, ok := genmsg.(*RelayStat)

	if !ok {
		t.Error("Failed to receive RelayStat")
		return
	}

	if msg.Nonce != "nonce" {
		t.Error("Failed to parse nonce")
	}

	if msg.Status != 0 {
		t.Error("Failed to parse status")
	}

	outgoing := genmsg.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}

func TestRelayAck(t *testing.T) {
	incoming := []byte("ACK nonce 0")
	_, genmsg, err := LoadMessage("relay", incoming)

	if err != nil {
		t.Error("Failed to create RelayAck")
		return
	}

	msg, ok := genmsg.(*RelayAck)

	if !ok {
		t.Error("Failed to receive RelayAck")
		return
	}

	if msg.Nonce != "nonce" {
		t.Error("Failed to parse nonce")
	}

	if msg.Status != 0 {
		t.Error("Failed to parse status")
	}

	outgoing := genmsg.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}


func TestRelayConf(t *testing.T) {
	incoming := []byte("CONF nonce 0")
	_, genmsg, err := LoadMessage("relay", incoming)

	if err != nil {
		t.Error("Failed to create RelayConf")
		return
	}

	msg, ok := genmsg.(*RelayConf)

	if !ok {
		t.Error("Failed to receive RelayConf")
		return
	}

	if msg.Nonce != "nonce" {
		t.Error("Failed to parse nonce")
	}

	if msg.Status != 0 {
		t.Error("Failed to parse status")
	}

	outgoing := genmsg.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}
