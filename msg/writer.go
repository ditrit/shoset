package msg

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"sync"
)

// Writer : simple bufio.Writer safe for goroutines...
type Writer struct {
	b *bufio.Writer
	m sync.Mutex
}

// NewWriter : constructor
func NewWriter(wd io.Writer) *Writer {
	s := new(Writer)
	s.b = bufio.NewWriter(wd)
	return s
}

// WriteString : safe version for goroutines
func (r *Writer) WriteString(data string) (int, error) {
	if r.b != nil {
		r.m.Lock()
		defer r.m.Unlock()
		return r.b.WriteString(data + "\n")
	}
	return 0, errors.New("Writer not ready")
}

// Flush : safe version for goroutines
func (r *Writer) Flush() error {
	if r.b != nil {
		r.m.Lock()
		defer r.m.Unlock()
		return r.b.Flush()
	}
	return errors.New("Writer not initialized")
}

// WriteMessage : encode a message in a safe way for goroutines
func (r *Writer) WriteMessage(data interface{}) error {
	if r.b != nil {
		r.m.Lock()
		defer r.m.Unlock()
		enc := gob.NewEncoder(r.b)
		err := enc.Encode(data)
		r.b.Flush()
		if err != nil {
			fmt.Printf("error in Writing Message : %s\n", err)
		}
		return err
	}
	return errors.New("Writer not initialized")
}
