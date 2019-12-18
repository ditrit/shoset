package main

import (
	"bufio"
	"encoding/gob"
	"io"
	"sync"
)

// SafeReader : simple bufio.Reader safe for goroutines...
type SafeReader struct {
	b *bufio.Reader
	m sync.Mutex
}

// NewSafeReader : constructor
func NewSafeReader(rd io.Reader) *SafeReader {
	s := new(SafeReader)
	s.b = bufio.NewReader(rd)
	return s
}

// ReadString : safe version for goroutines
func (r *SafeReader) ReadString() (string, error) {
	r.m.Lock()
	defer r.m.Unlock()
	return r.b.ReadString('\n')
}

// ReadMessage : encode a message in a safe way for goroutines
func (r *SafeReader) ReadMessage(data *interface{}) error {
	r.m.Lock()
	defer r.m.Unlock()
	enc := gob.NewDecoder(r.b)
	return enc.Decode(&data)
}

// SafeWriter : simple bufio.Writer safe for goroutines...
type SafeWriter struct {
	b *bufio.Writer
	m sync.Mutex
}

// NewSafeWriter : constructor
func NewSafeWriter(wd io.Writer) *SafeWriter {
	s := new(SafeWriter)
	s.b = bufio.NewWriter(wd)
	return s
}

// WriteString : safe version for goroutines
func (r *SafeWriter) WriteString(data string) (int, error) {
	r.m.Lock()
	defer r.m.Unlock()
	return r.b.WriteString(data + "\n")
}

// Flush : safe version for goroutines
func (r *SafeWriter) Flush() error {
	r.m.Lock()
	defer r.m.Unlock()
	return r.b.Flush()
}

// WriteMessage : encode a message in a safe way for goroutines
func (r *SafeWriter) WriteMessage(data *interface{}) error {
	r.m.Lock()
	defer r.m.Unlock()
	enc := gob.NewEncoder(r.b)
	return enc.Encode(&data)
}
