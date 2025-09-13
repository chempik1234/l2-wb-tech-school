package main

type StringNode struct {
	value string
	next  *StringNode
}

type Queue struct {
	head *StringNode
	tail *StringNode
	cap  int
	len  int
}

func NewQueue(cap int) *Queue {
	return &Queue{cap: cap}
}

func (l *Queue) Len() int {
	return l.len
}

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

func (l *Queue) RemoveFirst() {
	if l.head != nil {
		l.head = l.head.next
		l.len--
	}
}

func (l *Queue) Clear() {
	l.head = nil
	l.tail = nil
	l.len = 0
}

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
