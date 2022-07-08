package msg

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"sync"
	"syscall"
)

// Writer : simple bufio.Writer safe for goroutines...
type Writer struct {
	b   *bufio.Writer
	enc *gob.Encoder
	m   sync.Mutex
}

// UpdateWriter updates writer object with new connection information.
func (w *Writer) UpdateWriter(wd io.Writer) {
	w.m.Lock()
	defer w.m.Unlock()
	w.b = bufio.NewWriter(wd)
	w.enc = gob.NewEncoder(w.b)
}

// Flush writes any buffered data to the underlying io.Writer in a safe version for goroutines.
func (w *Writer) Flush() error {
	if w.b != nil {
		return w.b.Flush()
	}
	return errors.New("Writer not initialized")
}

// SendMessage send a message through a connection.
// Writes message type first.
// Then writes message.
func (w *Writer) SendMessage(msg Message) error {
	w.m.Lock()
	defer w.m.Unlock()

	_, err := w.WriteString(msg.GetMessageType())
	if err != nil {
		if errors.Is(err, syscall.EPIPE) {
			// https://gosamples.dev/broken-pipe/
			return nil
		} else if errors.Is(err, syscall.ECONNRESET) {
			// https://gosamples.dev/connection-reset-by-peer/
			return nil
		}
		return err
	}

	err = w.WriteMessage(msg)
	if err != nil {
		if errors.Is(err, syscall.EPIPE) {
			return nil
		} else if errors.Is(err, syscall.ECONNRESET) {
			return nil
		}
		return err
	}
	return nil
}

// WriteString writes a string in a safe version for goroutines.
func (w *Writer) WriteString(data string) (int, error) {
	if w.b != nil {
		return w.b.WriteString(data + "\n")
	}
	return 0, errors.New("Writer not ready")
}

// WriteMessage encodes a message in a safe way for goroutines.
func (w *Writer) WriteMessage(data interface{}) error {
	if w.b == nil {
		return errors.New("Writer not initialized")
	}

	err := w.enc.Encode(data)
	if err != nil {
		return err
	}
	return w.Flush()
}
