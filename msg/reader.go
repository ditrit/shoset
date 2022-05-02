package msg

import (
	"bufio"
	"encoding/gob"
	"github.com/rs/zerolog/log"
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

func (r *Reader) UpdateReader(rd io.Reader) {
	r.m.Lock()
	defer r.m.Unlock()
	r.b = bufio.NewReader(rd)
}

// ReadString : safe version for goroutines
func (r *Reader) ReadString() (string, error) {
	r.m.Lock()
	defer r.m.Unlock()
	return r.b.ReadString('\n')
}

// ReadMessage : decode a message in a safe way for goroutines
func (r *Reader) ReadMessage(data interface{}) error {
	r.m.Lock()
	defer r.m.Unlock()
	enc := gob.NewDecoder(r.b)
	err := enc.Decode(data)
	if err != nil {
		log.Error().Msg("error in ReadMessage : " + err.Error())
	}
	return err
}
