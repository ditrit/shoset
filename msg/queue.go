package msg

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

//Queue : queue allowing access via a string key
type Queue struct {
	qlist list.List
	dict  map[string]*list.Element
	m     sync.Mutex
}

// NewQueue : constructor
func NewQueue() *Queue {
	q := new(Queue)
	q.Init()
	return q
}

// Init :
func (q *Queue) Init() {
	q.qlist.Init()
	q.dict = make(map[string]*list.Element)
}

// Push : insert a new value in the queue except if the UUID is already present and remove after timeout expiration
func (q *Queue) Push(m Message) *list.Element {
	fmt.Printf("Push a message!")
	key := m.GetUUID()
	timeout := m.GetTimeout()
	q.m.Lock()
	defer q.m.Unlock()
	ele := q.dict[key]
	if ele != nil {
		return ele
	}
	ele = q.qlist.PushFront(m)
	q.dict[key] = ele
	go func() {
		time.Sleep(time.Duration(timeout) * time.Millisecond)
		q.removeByKey(key)
	}()
	return ele
}

// Get :
func (q *Queue) Get(ctx *list.Element) Message {
	q.m.Lock()
	defer q.m.Unlock()
	if ctx == nil {
		ctx = q.qlist.Back()
	} else {
		ctx = ctx.Prev()
	}
	if ctx == nil {
		return nil
	}
	data := ctx.Value.(Message)
	return data
}

// removeByKey :
func (q *Queue) removeByKey(key string) {
	q.m.Lock()
	defer q.m.Unlock()
	ele := q.dict[key]
	delete(q.dict, key)
	q.qlist.Remove(ele)
}

// Remove :
func (q *Queue) Remove(v Message) {
	key := v.GetUUID()
	q.removeByKey(key)
}

// IsEmpty : the event queue is empty
func (q *Queue) IsEmpty() bool {
	return q.qlist.Len() == 0
}
