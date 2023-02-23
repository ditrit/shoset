package msg

import (
	"sync"
)

//Iterator : queue allowing access via a string key
type Iterator struct {
	queue *Queue
	//seen    map[string]bool
	current string
	m       sync.Mutex
}

// NewIterator : constructor
func NewIterator(queue *Queue) *Iterator {
	i := new(Iterator)
	i.Init(queue)

	i.queue.m.Lock() //
	defer i.queue.m.Unlock() ///

	queue.iters[i] = true
	return i
}

// Init : initialisation
func (i *Iterator) Init(queue *Queue) {
	i.queue = queue
	//i.seen = make(map[string]bool)
}

// Close : fermeture de l'iterateur
func (i *Iterator) Close() {
	delete(i.queue.iters, i)
}

// Get : get next unseen element
func (i *Iterator) Get() *Cell {
	i.m.Lock()
	defer i.m.Unlock()

	var cell *Cell
	// Si la queue est vide, on ne renvoie rien
	if i.queue.IsEmpty() {
		return nil
	}

	// Si l'iterateur n'a pas été initialisé,
	if i.current == "" {
		cell = i.queue.First() // premiere cell de la queue
	} else {
		cell = i.queue.Next(i.current) // cell suivante
	}

	// si on a trouvé un nouveau message à renvoyer
	if cell != nil {
		i.current = (*cell).GetMessage().GetUUID() // on pointe dessus
	}
	return cell

	/*
		// on cherche une valeur qui n'a pas déjà été lue
		for i.seen[i.current] {
			message = i.queue.Next(i.current)
			if message == nil {
			return nil
			}
		}
	*/

	// la valeur a été vue
	//i.seen[i.current] = true

	//return message
}

// PrintQueue : print la queue
func (i *Iterator) PrintQueue() {
	i.queue.Print()
}

// PrintQueue : print la queue
func (i *Iterator) GetQueue() *Queue {
	return i.queue
}
