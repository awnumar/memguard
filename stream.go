package memguard

import (
	"container/list"
	"io"
	"sync"

	"github.com/awnumar/memguard/core"
)

// Stream is a streaming in-memory encrypted data vault.
type Stream struct {
	sync.Mutex
	*list.List
}

// NewStream initialises a new empty Stream object.
func NewStream() *Stream {
	s := new(Stream)
	s.List = list.New()
	return s
}

/*
Write encrypts and writes some given data to a Stream object. The last thing to be written to the Stream will be the last thing to be read.
*/ /* break up data int page-size chunks? */
func (s *Stream) Write(data []byte) (int, error) {
	s.Lock()
	defer s.Unlock()
	s.PushBack(NewEnclave(data))
	return len(data), nil
}

/*
Read decrypts and places some data from a Stream object into some provided buffer.
*/
func (s *Stream) Read(buf []byte) (int, error) {
	s.Lock()
	s.Unlock()

	// Pop data from the front of the list.
	e := s.Front()
	if e == nil {
		return 0, io.EOF
	}

	// Decrypt the data into a guarded allocation.
	b, err := e.Value.(*Enclave).Open()
	if err != nil {
		return 0, err
	}
	defer b.Destroy()

	// Copy the contents into the given buffer.
	core.Copy(buf, b.Bytes())

	// Check if there is data left over.
	if len(buf) < b.Size() {
		// Re-encrypt it and push onto the front of the list.
		n := b.Size() - len(buf)
		c := NewBuffer(n)
		c.Copy(b.Bytes()[n:])
		s.PushFront(c.Seal())
		return len(buf), nil
	}

	// Not enough data or perfect amount of data.
	// Either way we copied the entire buffer.
	return b.Size(), nil
}
