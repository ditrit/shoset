package msg

import (
	"bufio"
	"encoding/gob"
	"github.com/rs/zerolog/log"
	"io"
	"sync"
)

// Reader : simple bufio.Reader safe for goroutines.
type Reader struct {
	b *bufio.Reader // implements buffering for an io.Reader object
	dec *gob.Decoder // manages the receipt of type and data information read from the remote side of a connection. It is safe for concurrent use by multiple goroutines
	m sync.Mutex // mutex for goroutines synchronization
}

// UpdateReader updates reader object with new connection information.
func (r *Reader) UpdateReader(rd io.Reader) {
	r.m.Lock()
	defer r.m.Unlock()
	r.b = bufio.NewReader(rd)
	r.dec = gob.NewDecoder(r.b)
}

// ReadString reads until the first message in the input in a safe way for goroutines, returning a string containing the data.
func (r *Reader) ReadString() (string, error) {
	r.m.Lock()
	defer r.m.Unlock()
	return r.b.ReadString('\n')
}

// ReadMessage decodes a message in a safe way for goroutines.
func (r *Reader) ReadMessage(data interface{}) error {
	r.m.Lock()
	defer r.m.Unlock()
	err := r.dec.Decode(data)
	if err != nil {
		log.Error().Msg("error in ReadMessage : " + err.Error())
	}
	return err
}
