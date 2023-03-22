package fileSync

import (
	"container/list"
	"sync"

	"github.com/ditrit/shoset/msg"
)

/*
It implements a queue with a channel and lock functionnality.
It is used if FileTransfer to handle messages to send and to receive.
*/

type MessageQueue struct {
	qlist      list.List
	newMessage chan bool
	m          sync.Mutex
}

// NewMessageQueue : constructor
func NewMessageQueue() *MessageQueue {
	q := new(MessageQueue)
	q.qlist.Init()
	q.newMessage = make(chan bool, 1000)
	return q
}

func (q *MessageQueue) Push(m *msg.FileMessage) {
	q.m.Lock()
	q.qlist.PushBack(m)
	q.m.Unlock()
	q.newMessage <- true
}

func (q *MessageQueue) Pop() *msg.FileMessage {
	q.m.Lock()
	defer q.m.Unlock()
	e := q.qlist.Front()
	if e != nil {
		q.qlist.Remove(e)
		return e.Value.(*msg.FileMessage)
	}
	return nil
}

func (q *MessageQueue) GetChan() chan bool {
	return q.newMessage
}

// used by the piece structure for timeout to avoid reasking file we received
func (q *MessageQueue) HasChunk(begin int64) bool {
	q.m.Lock()
	defer q.m.Unlock()
	for e := q.qlist.Front(); e != nil; e = e.Next() {
		message := e.Value.(*msg.FileMessage)
		if message.Begin == begin {
			return true
		}
	}
	return false
}
