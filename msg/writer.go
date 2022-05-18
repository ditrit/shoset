package msg

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"sync"
)

// Writer : simple bufio.Writer safe for goroutines...
type Writer struct {
	b *bufio.Writer
	enc *gob.Encoder
	m sync.Mutex
}

func (w *Writer) GetBufioWriter() *bufio.Writer {
	w.m.Lock()
	defer w.m.Unlock()
	return w.b
}

func (r *Writer) UpdateWriter(wd io.Writer) {
	r.m.Lock()
	defer r.m.Unlock()
	r.b = bufio.NewWriter(wd)
	r.enc = gob.NewEncoder(r.b)
}

// Flush : safe version for goroutines
func (r *Writer) Flush() error {
	if r.b != nil {
		return r.b.Flush()
	}
	return errors.New("Writer not initialized")
}

// WriteString writes a string in a safe version for goroutines
func (r *Writer) WriteString(data string) (int, error) {
	if r.b != nil {
		r.m.Lock()
		defer r.m.Unlock()
		return r.b.WriteString(data + "\n")
	}
	return 0, errors.New("Writer not ready")
}


// WriteMessage : encode a message in a safe way for goroutines
func (r *Writer) WriteMessage(data interface{}) error {
	if r.b == nil {
		return errors.New("Writer not initialized")
	}

	r.m.Lock()
	defer r.m.Unlock()
	err := r.enc.Encode(data)
	if err != nil {
		return err
	}
	return r.Flush()
}
