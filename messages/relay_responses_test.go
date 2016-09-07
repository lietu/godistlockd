package messages

import (
	"testing"
	"bytes"
)

func TestRelayOutgoingHello(t *testing.T) {
	expected := []byte("HOWDY nonce server-1 1.0.0")

	msg, err := NewRelayHowdy([]string{"nonce", "server-1", "1.0.0"})

	if err != nil {
		t.Error("Failed to create RelayHowdy")
	}

	outgoing := msg.ToBytes()
	if !bytes.Equal(outgoing, expected) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}
