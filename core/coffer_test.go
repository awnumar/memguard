package core

import (
	"bytes"
	"math/rand/v2"
	"sync"
	"testing"
	"time"
)

func TestNewCoffer(t *testing.T) {
	s := NewCoffer()

	// Attain a lock to halt the verify & rekey cycle.
	s.Lock()

	// Verify that fields are not nil.
	if s.left == nil || s.right == nil {
		t.Error("one or more fields are not initialised")
	}

	// Verify that fields are the expected sizes.
	if len(s.left.Data()) != 32 {
		t.Error("left side has unexpected lengths")
	}
	if len(s.right.Data()) != 32 {
		t.Error("right size has unexpected lengths")
	}

	// Verify that the data fields are not zeroed.
	if bytes.Equal(s.left.Data(), make([]byte, 32)) {
		t.Error("left side is zeroed")
	}
	if bytes.Equal(s.right.Data(), make([]byte, 32)) {
		t.Error("right side is zeroed")
	}

	s.Unlock() // Release mutex to allow destruction.
	s.Destroy()
}

func TestCofferInit(t *testing.T) {
	s := NewCoffer()

	// Get the value stored inside.
	view, err := s.View()
	if err != nil {
		t.Error("unexpected error")
	}
	value := make([]byte, 32)
	copy(value, view.Data())
	view.Destroy()

	// Re-init the buffer with a new value.
	if err := s.Init(); err != nil {
		t.Error("unexpected error;", err)
	}

	// Get the new value stored inside.
	view, err = s.View()
	if err != nil {
		t.Error("unexpected error")
	}
	newValue := make([]byte, 32)
	copy(newValue, view.Data())
	view.Destroy()

	// Compare them.
	if bytes.Equal(value, newValue) {
		t.Error("value was not refreshed")
	}

	s.Destroy()

	// Check error condition.
	if err := s.Init(); err != ErrCofferExpired {
		t.Error("expected ErrCofferExpired; got", err)
	}
}

func TestCofferView(t *testing.T) {
	s := NewCoffer()

	// Get the value stored inside.
	view, err := s.View()
	if err != nil {
		t.Error("unexpected error")
	}
	if view == nil {
		t.Error("returned object is nil")
	}

	// Some sanity checks on the inner value.
	if view.Data() == nil || len(view.Data()) != 32 {
		t.Error("unexpected data; got", view.Data())
	}
	if bytes.Equal(view.Data(), make([]byte, 32)) {
		t.Error("value inside coffer is zero")
	}

	// Destroy our temporary view of the coffer's contents.
	view.Destroy()

	s.Destroy()

	// Check error condition.
	view, err = s.View()
	if err != ErrCofferExpired {
		t.Error("expected ErrCofferExpired; got", err)
	}
	if view != nil {
		t.Error("expected nil buffer object")
	}
}

func TestCofferRekey(t *testing.T) {
	s := NewCoffer()

	// remember the value stored inside
	view, err := s.View()
	if err != nil {
		t.Error("unexpected error;", err)
	}
	orgValue := make([]byte, 32)
	copy(orgValue, view.Data())
	view.Destroy()

	// remember the value of the partitions
	left := make([]byte, 32)
	right := make([]byte, 32)
	s.Lock() // halt re-key cycle
	copy(left, s.left.Data())
	copy(right, s.right.Data())
	s.Unlock() // un-halt re-key cycle

	s.Rekey() // force a re-key

	view, err = s.View()
	if err != nil {
		t.Error("unexpected error;", err)
	}
	newValue := make([]byte, 32)
	copy(newValue, view.Data())
	view.Destroy()

	if !bytes.Equal(orgValue, newValue) {
		t.Error("value inside coffer changed!!")
	}

	if bytes.Equal(left, s.left.Data()) || bytes.Equal(right, s.right.Data()) {
		t.Error("partition values did not change")
	}

	s.Destroy()

	if err := s.Rekey(); err != ErrCofferExpired {
		t.Error("expected ErrCofferExpired; got", err)
	}
}

func TestCofferDestroy(t *testing.T) {
	s := NewCoffer()
	s.Destroy()

	// Check metadata flags.
	if !s.Destroyed() {
		t.Error("expected destroyed")
	}

	// Check both partitions are destroyed.
	if s.left.alive || s.right.alive {
		t.Error("some partition not destroyed")
	}
}

func TestCofferConcurrent(t *testing.T) {
	testConcurrency := 4
	testDuration := 1 * time.Second

	funcs := []func(s *Coffer) error{
		func(s *Coffer) error {
			return s.Rekey()
		},
		func(s *Coffer) error {
			_, err := s.View()
			return err
		},
	}
	wg := &sync.WaitGroup{}

	s := NewCoffer()
	defer s.Destroy()

	start := time.Now()

	for range testConcurrency {
		wg.Add(1)
		go func(t *testing.T) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					// Log panic -- it's likely just ran out of mlock space.
					t.Logf("Recovered from panic: %s", r)
				}
			}()
			fIndex := rand.IntN(len(funcs))
			for time.Since(start) < testDuration {
				err := funcs[fIndex](s)
				if err != nil && err != ErrCofferExpired {
					t.Errorf("unexpected error: %v", err)
				}
			}
		}(t)
	}

	wg.Wait()
}
