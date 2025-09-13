package main

// StringNode is a node for Queue
type StringNode struct {
	value string
	next  *StringNode
}

// Queue is a FIFO linked buffer
type Queue struct {
	head *StringNode
	tail *StringNode
	cap  int
	len  int
}

// NewQueue creates an empty Queue with given cap
func NewQueue(cap int) *Queue {
	return &Queue{cap: cap}
}

// Len returns the length of a Queue
func (l *Queue) Len() int {
	return l.len
}

// Append adds the item to the end of the queue
func (l *Queue) Append(value string) {
	if l.cap == 0 {
		return
	}

	for l.len >= l.cap {
		l.RemoveFirst()
	}

	newNode := &StringNode{value, nil}

	if l.head == nil {
		l.head = newNode
	}

	if l.tail != nil {
		l.tail.next = newNode
	}
	l.tail = newNode

	l.len++
}

// RemoveFirst removes the first item from queue
func (l *Queue) RemoveFirst() {
	if l.head != nil {
		l.head = l.head.next
		l.len--
	}
}

// Clear clears the queue by resetting pointers to first/last elem and len=0
func (l *Queue) Clear() {
	l.head = nil
	l.tail = nil
	l.len = 0
}

// All returns a channel that values must be read from
func (l *Queue) All() <-chan string {
	ch := make(chan string)

	go func() {
		currentElem := l.head
		for currentElem != nil {
			ch <- currentElem.value
			currentElem = currentElem.next
		}
		close(ch)
	}()

	return ch
}
