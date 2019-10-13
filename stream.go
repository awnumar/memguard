package memguard

import (
	"container/list"
	"io"
	"os"
	"sync"

	"github.com/awnumar/memguard/core"
)

type queue struct {
	sync.Mutex
	*list.List
}

// add data to back of queue
func (q *queue) push(e *Enclave) {
	q.Lock()
	q.PushBack(e)
	q.Unlock()
}

// pop data off front of queue
// returns nil if queue is empty
func (q *queue) pop() *Enclave {
	q.Lock()
	defer q.Unlock()

	e := q.Front() // get element at front of queue
	if e == nil {
		return nil // no data
	}
	q.Remove(e)               // success => remove value
	return e.Value.(*Enclave) // unwrap and return (potential panic)
}

// Stream is a streaming in-memory encrypted data vault.
type Stream struct {
	*queue
}

// NewStream initialises a new empty Stream object.
func NewStream() *Stream {
	return &Stream{queue: &queue{List: list.New()}}
}

/*
Write encrypts and writes some given data to a Stream object. The last thing to be written to the Stream will be the last thing to be read.
*/
func (s *Stream) Write(data []byte) (int, error) {
	ps := os.Getpagesize()
	for i := 0; i < len(data); i += ps {
		if i+ps > len(data) {
			s.push(NewEnclave(data[len(data)-(len(data)%ps):]))
		} else {
			s.push(NewEnclave(data[i : i+ps]))
		}
	}
	return len(data), nil
}

/*
Read decrypts and places some data from a Stream object into some provided buffer.
*/
func (s *Stream) Read(buf []byte) (int, error) {
	s.Lock()
	s.Unlock()

	// Pop data from the front of the list.
	e := s.pop()
	if e == nil {
		return 0, io.EOF
	}

	// Decrypt the data into a guarded allocation.
	b, err := e.Open()
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
		s.push(c.Seal())
		return len(buf), io.ErrShortBuffer
	}

	// Not enough data or perfect amount of data.
	// Either way we copied the entire buffer.
	return b.Size(), nil
}
