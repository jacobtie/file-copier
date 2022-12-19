package ds

import (
	"fmt"
	"sync"
)

type Node[T any] struct {
	Prev    *Node[T]
	Next    *Node[T]
	Element T
}

type ThreadSafeQueue[T any] struct {
	First *Node[T]
	Last  *Node[T]
	Count int
	mutex *sync.Mutex
}

func NewThreadSafeQueue[T any](elems ...T) *ThreadSafeQueue[T] {
	q := &ThreadSafeQueue[T]{
		First: nil,
		Last:  nil,
		Count: 0,
		mutex: &sync.Mutex{},
	}
	if len(elems) > 0 {
		q.Enqueue(elems...)
	}
	return q
}

func (q *ThreadSafeQueue[T]) Enqueue(elems ...T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for _, elem := range elems {
		q.enqueueOne(elem)
	}
}

func (q *ThreadSafeQueue[T]) enqueueOne(elem T) {
	n := &Node[T]{
		Prev:    nil,
		Next:    nil,
		Element: elem,
	}
	if q.IsEmpty() {
		q.First = n
		q.Last = n
		q.Count = 1
		return
	}
	q.Last.Next = n
	n.Prev = q.Last
	q.Last = n
	q.Count++
}

func (q *ThreadSafeQueue[T]) Dequeue() (T, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if q.IsEmpty() {
		return *new(T), fmt.Errorf("queue is empty")
	}
	elem := q.First.Element
	if q.Count == 1 {
		q.First = nil
		q.Last = nil
		q.Count = 0
		return elem, nil
	}
	q.First.Next.Prev = nil
	q.First = q.First.Next
	q.Count--
	return elem, nil
}

func (q ThreadSafeQueue[T]) IsEmpty() bool {
	return q.Count == 0
}
