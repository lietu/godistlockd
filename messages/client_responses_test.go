package messages

import (
	"testing"
	"bytes"
)

func TestClientOutgoingGive(t *testing.T) {
	expected := []byte("GIVE nonce fence")

	msg := NewClientOutgoingGive("nonce", "fence")
	outgoing := msg.ToBytes()
	if !bytes.Equal(outgoing, expected) {
		t.Error("Failed to convert back to bytes:", string(outgoing))
	}
}
