package msg

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
)

// Reader : simple bufio.Reader safe for goroutines...
type Reader struct {
	b *bufio.Reader
	m sync.Mutex
}

// NewReader : constructor
func NewReader(rd io.Reader) *Reader {
	s := new(Reader)
	s.b = bufio.NewReader(rd)
	return s
}

// ReadString : safe version for goroutines
func (r *Reader) ReadString() (string, error) {
	r.m.Lock()
	defer r.m.Unlock()
	return r.b.ReadString('\n')
}

// ReadEvent : encode a message in a safe way for goroutines
func (r *Reader) ReadEvent(data *Event) error {
	r.m.Lock()
	defer r.m.Unlock()
	enc := gob.NewDecoder(r.b)
	err := enc.Decode(&data)
	if err != nil {
		fmt.Printf("error in ReadEvent : %s\n", err)
	}
	return err
}

// ReadCommand : encode a message in a safe way for goroutines
func (r *Reader) ReadCommand(data *Command) error {
	r.m.Lock()
	defer r.m.Unlock()
	enc := gob.NewDecoder(r.b)
	return enc.Decode(data)
}

// ReadReply : encode a message in a safe way for goroutines
func (r *Reader) ReadReply(data *Reply) error {
	r.m.Lock()
	defer r.m.Unlock()
	enc := gob.NewDecoder(r.b)
	return enc.Decode(data)
}

// ReadConfig : encode a message in a safe way for goroutines
func (r *Reader) ReadConfig(data *Config) error {
	r.m.Lock()
	defer r.m.Unlock()
	enc := gob.NewDecoder(r.b)
	return enc.Decode(data)
}
