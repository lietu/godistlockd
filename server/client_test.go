package server

import (
	"testing"
)

func TestHeldLockCleanup(t *testing.T) {
	c := NewClient(nil, nil)
	c.addLock("foo")
	c.removeLock("foo")

	if len(c.GetHeldLocks()) != 0 {
		t.Error("Held locks left lingering")
	}
}