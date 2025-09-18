package main

import (
	"io"
	"sync"
)

// ChanReader is an io.Reader implementation for <-chan string
type ChanReader struct {
	ch     <-chan string
	buffer string
	mu     sync.Mutex
}

// NewChannelReader creates a new ChanReader with given chan to read
func NewChannelReader(ch <-chan string) *ChanReader {
	return &ChanReader{
		ch: ch,
	}
}

// Read method smartly returns needed data, fetching strings from channel on demand
func (r *ChanReader) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// as strings aren't used fully, we store them to finish later
	// 1) if there is some buffer we return as much as we can put into p
	// 2) if buffer is empty then create it lol and move to step 1
	if len(r.buffer) == 0 {
		s, ok := <-r.ch // if we can't read, it's EOF
		if !ok {
			return 0, io.EOF
		}
		r.buffer = s
	}

	n = copy(p, r.buffer)
	r.buffer = r.buffer[n:]

	return n, nil
}
