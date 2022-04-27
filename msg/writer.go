package msg

import (
	"bufio"
	"encoding/gob"
	"errors"

	// "fmt"
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

func (r *Writer) UpdateWriter(wd io.Writer) {
	r.m.Lock()
	defer r.m.Unlock()
	r.b = bufio.NewWriter(wd)
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
		return r.b.Flush()
	}
	return errors.New("Writer not initialized")
}

// WriteMessage : encode a message in a safe way for goroutines
func (r *Writer) WriteMessage(data interface{}) error {
	if r.b == nil {
		return errors.New("Writer not initialized")
	}

	r.m.Lock()
	defer r.m.Unlock()
	enc := gob.NewEncoder(r.b)
	err := enc.Encode(data)
	if err != nil {
		// data2 := data.(Message)
		// if data2.GetMsgType() == "cfglink" {
		// 	linkProtocol := data2.(*ConfigProtocol)
		// 	fmt.Println("-------")
		// 	fmt.Println(linkProtocol.GetCommandName())
		// 	fmt.Println(linkProtocol.GetAddress())
		// 	fmt.Println("-------")
		// }
		// fmt.Println(data2.GetMsgType())
		// fmt.Printf("error in Writing Message : %s\n", err)
		return err
	}
	return r.Flush()
}
