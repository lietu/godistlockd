package messages

import (
	"testing"
	"bytes"
)

func TestRelayIncomingHello(t *testing.T) {
	incoming := []byte("HELLO server-1 1.0.0 mynonce")
	_, msg, err := LoadMessage("relay", incoming)

	if err != nil {
		t.Error("Failed to parse RelayIncomingHello")
		return
	}

	rih, ok := msg.(*RelayIncomingHello)

	if !ok {
		t.Error("Failed to receive RelayIncomingHello")
		return
	}

	if rih.Id != "server-1" {
		t.Error("Failed to parse id")
	}

	if rih.Version != "1.0.0" {
		t.Error("Failed to parse version")
		return
	}

	if rih.Nonce != "mynonce" {
		t.Error("Failed to parse nonce")
		return
	}

	outgoing := rih.ToBytes()
	if !bytes.Equal(outgoing, incoming) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}
