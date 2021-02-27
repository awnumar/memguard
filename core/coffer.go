package core

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Interval of time between each verify & re-key cycle.
const interval = 500 * time.Millisecond

// ErrCofferExpired is returned when a function attempts to perform an operation using a secure key container that has been wiped and destroyed.
var ErrCofferExpired = errors.New("<memguard::core::ErrCofferExpired> attempted usage of destroyed key object")

var key = func() *Coffer {
	s, err := NewCoffer()
	if err != nil {
		panic(err)
	}
	return s
}()

/*
Coffer is a specialized container for securing highly-sensitive, 32 byte values.
*/
type Coffer struct {
	mu sync.RWMutex // caller's responsibility to acquire

	left  *Buffer // Left partition.
	right *Buffer // Right partition.

	rand *Buffer // Static allocation for fast random bytes reading.

	// Setting this instructs the rekey routine to return.
	// You must also destroy the partitions.
	discarded uint32
}

// NewCoffer is a raw constructor for the *Coffer object.
func NewCoffer() (s *Coffer, err error) {
	s = &Coffer{}

	s.left, err = NewBuffer(32)
	if err != nil {
		return
	}

	s.right, err = NewBuffer(32)
	if err != nil {
		return
	}

	s.rand, err = NewBuffer(32)
	if err != nil {
		return
	}

	err = s.init()
	if err != nil {
		return
	}

	go func(s *Coffer) {
		var err error
		for {
			time.Sleep(interval)

			if atomic.LoadUint32(&s.discarded) == 1 {
				return
			}

			s.mu.Lock()
			err = s.rekey()
			s.mu.Unlock()

			if err != nil {
				fmt.Println()
				return
			}
		}
	}(s)

	return
}

func (s *Coffer) init() error {
	if s.destroyed() {
		return ErrCofferExpired
	}

	if err := Scramble(s.left.Data()); err != nil {
		return err
	}
	if err := Scramble(s.right.Data()); err != nil {
		return err
	}

	// left = left XOR hash(right)
	hr := Hash(s.right.Data())
	for i := range hr {
		s.left.Data()[i] ^= hr[i]
	}
	Wipe(hr)

	return nil
}

func (s *Coffer) view() (b *Buffer, err error) {
	if s.destroyed() {
		err = ErrCofferExpired
		return
	}

	b, err = NewBuffer(32)
	if err != nil {
		return
	}

	// data = hash(right) XOR left
	h := Hash(s.right.Data())

	for i := range b.Data() {
		b.Data()[i] = h[i] ^ s.left.Data()[i]
	}

	Wipe(h)

	return
}

func (s *Coffer) rekey() error {
	if s.destroyed() {
		return ErrCofferExpired
	}

	// Attain 32 bytes of fresh cryptographic buf32.
	if err := Scramble(s.rand.Data()); err != nil {
		return err
	}

	hashRightCurrent := Hash(s.right.Data())

	// new_right = current_right XOR buf32
	for i := range s.right.Data() {
		s.right.Data()[i] ^= s.rand.Data()[i]
	}

	// new_left = current_left XOR hash(current_right) XOR hash(new_right)
	hashRightNew := Hash(s.right.Data())
	for i := range s.left.Data() {
		s.left.Data()[i] ^= hashRightCurrent[i] ^ hashRightNew[i]
	}
	Wipe(hashRightNew)

	return nil
}

func (s *Coffer) destroy() error {
	if s.destroyed() {
		return nil
	}

	err1 := s.left.Destroy()
	err2 := s.right.Destroy()
	err3 := s.rand.Destroy()

	errS := ""
	if err1 != nil {
		errS = errS + "Error destroying left partition: " + err1.Error() + "\n"
	}
	if err2 != nil {
		errS = errS + "Error destroying right partition: " + err2.Error() + "\n"
	}
	if err3 != nil {
		errS = errS + "Error destroying rand partition: " + err3.Error() + "\n"
	}
	if errS == "" {
		return nil
	}
	return errors.New(errS)
}

func (s *Coffer) destroyed() bool {
	return !s.left.Alive() && !s.right.Alive()
}
