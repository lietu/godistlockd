package server

import (
	"testing"
	"time"
)

func TestLockManagerBasics(t *testing.T) {
	lm := NewLockManager()

	got := false

	go func() {
		lm.GetLock("id", "foo", time.Millisecond * 100)
		got = true
	}()

	time.Sleep(time.Millisecond * 25)

	if !got {
		t.Error("Failed to get lock within reasonable time")
	}

	lm.Stop()
}

func TestLockManagerGetLock(t *testing.T) {
	lm := NewLockManager()

	lm.GetLock("id", "foo", time.Millisecond * 100)

	got := false
	go func() {
		lm.GetLock("id2", "foo", time.Millisecond * 100)
		got = true
	}()

	time.Sleep(time.Millisecond * 25)

	if got {
		t.Error("Got lock before timeout expired")
	}

	time.Sleep(time.Millisecond * 100)

	if !got {
		t.Error("Failed to get lock after timeout expired")
	}

	lm.Stop()
}

func TestLockManagerTryGet(t *testing.T) {
	lm := NewLockManager()

	lm.GetLock("id", "foo", time.Millisecond * 100)

	res := lm.TryGet("id2", "foo", time.Millisecond * 100)

	if res != nil {
		t.Error("TryGet succeeded while lock was being held")
	}

	time.Sleep(time.Millisecond * 100)

	res = lm.TryGet("id", "foo", time.Millisecond * 100)

	if res == nil {
		t.Error("TryGet failed after lock timed out")
	}

	lm.Stop()
}

func TestLockManagerIsLocked(t *testing.T) {
	lm := NewLockManager()

	res := lm.IsLocked("foo")

	if res != "" {
		t.Error("IsLocked before locking")
	}

	lm.GetLock("id", "foo", time.Millisecond * 100)

	res = lm.IsLocked("foo")
	if res == "" {
		t.Error("IsLocked didn't return fence after locking")
	}

	lm.Stop()
}

func TestLockManagerRelock(t *testing.T) {
	lm := NewLockManager()

	lm.GetLock("id", "foo", time.Millisecond * 10000000)
	got := false

	go func() {
		lm.GetLock("id", "foo", time.Millisecond * 100000)
		got = true
	}()

	time.Sleep(time.Millisecond * 25)

	if !got {
		t.Error("Failed to re-acquire lock for same client")
	}

	lm.Stop()
}
