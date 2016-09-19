package server

import (
	"testing"
	"fmt"
)

func TestRelayManagerCalculateQuorum(t *testing.T) {
	quorum := calculateQuorum(1)
	if quorum != 1 {
		t.Error(fmt.Sprintf("Quorum for 1 server should be 1, not %d", quorum))
	}

	quorum = calculateQuorum(2)
	if quorum != 2 {
		t.Error(fmt.Sprintf("Quorum for 2 servers should be 2, not %d", quorum))
	}

	quorum = calculateQuorum(3)
	if quorum != 2 {
		t.Error(fmt.Sprintf("Quorum for 3 servers should be 2, not %d", quorum))
	}

	quorum = calculateQuorum(4)
	if quorum != 3 {
		t.Error(fmt.Sprintf("Quorum for 4 servers should be 3, not %d", quorum))
	}

	quorum = calculateQuorum(5)
	if quorum != 3 {
		t.Error(fmt.Sprintf("Quorum for 5 servers should be 3, not %d", quorum))
	}

	quorum = calculateQuorum(6)
	if quorum != 4 {
		t.Error(fmt.Sprintf("Quorum for 6 servers should be 4, not %d", quorum))
	}

	quorum = calculateQuorum(7)
	if quorum != 4 {
		t.Error(fmt.Sprintf("Quorum for 7 servers should be 4, not %d", quorum))
	}
}