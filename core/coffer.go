package core

import (
	"errors"
	"sync"
	"time"
)

// Interval of time between each verify & re-key cycle.
const interval = 500 * time.Millisecond

// ErrCofferExpired is returned when a function attempts to perform an operation using a secure key container that has been wiped and destroyed.
var ErrCofferExpired = errors.New("<memguard::core::ErrCofferExpired> attempted usage of destroyed key object")

/*
Coffer is a specialized container for securing highly-sensitive, 32 byte values.
*/
type Coffer struct {
	sync.RWMutex

	left  *Buffer // Left partition.
	right *Buffer // Right partition.

	rand *Buffer // Static allocation for fast random bytes reading.
}

// NewCoffer is a raw constructor for the *Coffer object.
func NewCoffer() *Coffer {
	s := new(Coffer)

	s.left, _ = NewBuffer(32)
	s.right, _ = NewBuffer(32)
	s.rand, _ = NewBuffer(32)

	s.Initialise()

	go func(s *Coffer) {
		for {
			time.Sleep(interval)

			// Re-key the contents, exiting the routine if object destroyed.
			if err := s.rekey(); err != nil {
				break
			}
		}
	}(s)

	return s
}

/*
Initialise is used to reset the value stored inside a Coffer to a new random 32 byte value, overwriting the old.
*/
func (s *Coffer) Initialise() {
	s.Lock()
	defer s.Unlock()

	if err := s.initialise(); err != nil {
		Panic(err)
	}
}

func (s *Coffer) initialise() error {
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

/*
View returns a snapshot of the contents of a Coffer inside a Buffer. As usual the Buffer should be destroyed as soon as possible after use by calling the Destroy method.
*/
func (s *Coffer) View() (*Buffer, error) {
	s.Lock()
	defer s.Unlock()

	return s.view()
}

func (s *Coffer) view() (*Buffer, error) {
	if s.destroyed() {
		return nil, ErrCofferExpired
	}

	b, _ := NewBuffer(32)

	// data = hash(right) XOR left
	h := Hash(s.right.Data())

	for i := range b.Data() {
		b.Data()[i] = h[i] ^ s.left.Data()[i]
	}

	Wipe(h)

	return b, nil
}

/*
Rekey is used to re-key a Coffer. Ideally this should be done at short, regular intervals.
*/
func (s *Coffer) Rekey() error {
	// Attain the mutex.
	s.Lock()
	defer s.Unlock()

	if err := s.rekey(); err != nil {
		Panic(err)
	}

	return nil
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

/*
Destroy wipes and cleans up all memory related to a Coffer object. Once this method has been called, the Coffer can no longer be used and a new one should be created instead.
*/
func (s *Coffer) Destroy() {
	// Attain the mutex.
	s.Lock()
	defer s.Unlock()

	if err := s.destroy(); err != nil {
		Panic(err)
	}
}

func (s *Coffer) destroy() error {
	if s.destroyed() {
		return nil
	}

	// Destroy the partitions.
	err1 := s.left.destroy()
	if err1 == nil {
		buffers.remove(s.left)
	}

	err2 := s.right.destroy()
	if err2 == nil {
		buffers.remove(s.right)
	}

	err3 := s.rand.destroy()
	if err3 == nil {
		buffers.remove(s.rand)
	}

	errS := ""
	if err1 != nil {
		errS = errS + err1.Error() + "\n"
	}
	if err2 != nil {
		errS = errS + err2.Error() + "\n"
	}
	if err3 != nil {
		errS = errS + err3.Error() + "\n"
	}
	if errS == "" {
		return nil
	}
	return errors.New(errS)
}

// Destroyed returns a boolean value indicating if a Coffer has been destroyed.
func (s *Coffer) Destroyed() bool {
	s.RLock()
	defer s.RUnlock()

	return s.destroyed()
}

func (s *Coffer) destroyed() bool {
	return !s.left.Alive() && !s.right.Alive()
}
