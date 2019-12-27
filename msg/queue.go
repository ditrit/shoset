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

// Back :
func (q *Queue) Back() *list.Element {
	q.m.Lock()
	defer q.m.Unlock()
	return q.qlist.Back()
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

// Print :
func (q *Queue) Print() {
	ele := q.qlist.Back()
	fmt.Printf("   Queue{\n")
	for ele != nil {
		fmt.Printf("      %s,\n", ele.Value.(Message).GetUUID())
		ele = ele.Prev()
	}
	fmt.Printf("nb eles : %d\n", q.qlist.Len())
}

//QueueIterator : queue allowing access via a string key
type QueueIterator struct {
	queue   *Queue
	seen    map[string]bool
	current *list.Element
	m       sync.Mutex
}

// NewQueueIterator : constructor
func NewQueueIterator(queue *Queue) *QueueIterator {
	i := new(QueueIterator)
	i.Init(queue)
	return i
}

// Init : initialisation
func (i *QueueIterator) Init(queue *Queue) {
	i.queue = queue
	i.seen = make(map[string]bool)
}

// Get : get next unread element
func (i *QueueIterator) Get() *Message {
	i.m.Lock()
	defer i.m.Unlock()
	for {
		if i.current == nil {
			// rewind !!
			i.current = i.queue.Back()
		}
		if i.current == nil {
			//queue is empty
			return nil
		}
		message := i.current.Value.(Message)
		uuid := message.GetUUID()
		seen := i.seen[uuid]
		if seen == false {
			i.seen[uuid] = true
			go func() {
				time.Sleep(time.Duration(message.GetTimeout()) * time.Millisecond)
				i.m.Lock()
				delete(i.seen, uuid)
				i.m.Unlock()
			}()
			return &message
		}
		i.current = i.current.Prev()
	}
}
