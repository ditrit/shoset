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
	iters map[*Iterator]bool
	m     sync.Mutex
}

// Cell : witch contain messages and useful intel
type Cell struct {
	key              string
	timeout          int64
	RemoteShosetType string
	RemoteAddress    string
	m                Message
}

// GetMessage :
func (c *Cell) GetMessage() Message { return c.m }

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
	q.iters = make(map[*Iterator]bool)
}

// Init :
func (q *Queue) GetByReferencesUUID(uuid string) *Event {
	//q.m.Lock()
	//defer q.m.Unlock()
	for _, val := range q.dict {
		value := val.Value.(Cell).m.(Event)
		if uuid == value.GetReferenceUUID() {
			return &value
		}
	}

	return nil
}

// Push : insert a new value in the queue except if the UUID is already present and remove after timeout expiration
func (q *Queue) Push(m Message, RemoteShosetType, RemoteAddress string) bool {
	// fmt.Printf("Push a message!\n")

	// Let's first initialize the Cell with all our data
	var c Cell
	c.key = m.GetUUID()
	// fmt.Println("key")
	// fmt.Println(c.key)
	c.timeout = m.GetTimeout()
	c.RemoteShosetType = RemoteShosetType
	c.RemoteAddress = RemoteAddress
	c.m = m

	q.m.Lock()
	defer q.m.Unlock()
	ele := q.dict[c.key]
	if ele != nil {
		return false
	}
  
	ele = q.qlist.PushFront(c)
	q.dict[c.key] = ele
	go func() {
		time.Sleep(time.Duration(c.timeout) * time.Millisecond)
		q.remove(c.key)
	}()
	return true
}

// First :
func (q *Queue) First() *Cell {
	q.m.Lock()
	defer q.m.Unlock()
	ele := q.qlist.Back()
	if ele != nil {
		value := ele.Value.(Cell)
		return &value
	}
	return nil
}

// Next :
func (q *Queue) Next(key string) *Cell {
	q.m.Lock()
	defer q.m.Unlock()
	cellFromKey := q.dict[key]
	if cellFromKey != nil {
		nextEle := cellFromKey.Prev()
		if nextEle != nil {
			nextMessage := nextEle.Value.(Cell)
			return &nextMessage
		}
	}
	return nil
}

// remove :
func (q *Queue) remove(key string) {
	q.m.Lock()
	defer q.m.Unlock()

	// Repositionner les iterateurs positionnés sur le message à supprimer
	// sur le message :
	// 1. suivant s'il existe
	// 2. sinon sur le précédent s'il existe
	// 3. sinon c'est que la queue est vide
	cell := q.dict[key]
	nextCell := cell.Prev() // cas 1.
	if nextCell == nil {
		nextCell = cell.Next() // cas 2.
	}
	nextUUID := ""
	if nextCell != nil {
		nextUUID = nextCell.Value.(Cell).m.GetUUID()
	}

	// suppression et repositionnement pour chaque iterateur
	for i := range q.iters {
		i.m.Lock()
		if i.current == key { // si literateur pointe sur le message à supprimer
			i.current = nextUUID // repositionnement
		}
		// supprimer le message de la liste des messages déjà consultés
		//delete(i.seen, key)
		i.m.Unlock()
	}

	// supprimer le message dans la queue (dans la liste et dans la map)
	delete(q.dict, key)
	q.qlist.Remove(cell)
}

// IsEmpty : the event queue is empty
func (q *Queue) IsEmpty() bool {
	return q.qlist.Len() == 0
}

// Print :
func (q *Queue) Print() {
	cell := q.qlist.Back()
	fmt.Printf("   Queue{\n")
	for cell != nil {
		fmt.Printf("      %s,\n", cell.Value.(Cell).m.GetUUID())
		cell = cell.Prev()
	}
	fmt.Printf("nb cell : %d\n", q.qlist.Len())
}
