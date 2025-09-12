package main

import (
	"sync/atomic"
)

type node[T any] struct {
	value T
	next  *node[T]
}

type linkedList[T any] struct {
	head *node[T]
	len  atomic.Int32
}

func newLinkedList[T any]() *linkedList[T] {
	return &linkedList[T]{}
}

// insertSorted inserts an element into linked list
func (l *linkedList[T]) insertSorted(value T, comparator func(a, b T) bool) {
	currentNode := l.head

	var valueIsLess bool
	var prevNode *node[T]
	for currentNode != nil {
		// value is still < cur elem
		valueIsLess = !comparator(value, currentNode.value)
		if !valueIsLess {
			break
		}

		prevNode = currentNode
		currentNode = currentNode.next
		if currentNode == nil {
			break
		}
	}

	l.len.Add(1)

	if prevNode == nil {
		l.head = &node[T]{value, l.head}
		return
	}
	prevNode.next = &node[T]{value, currentNode}
	return
}

func (l *linkedList[T]) build() []T {
	result := make([]T, 0, l.len.Load())
	currentNode := l.head
	for currentNode != nil {
		result = append(result, currentNode.value)
		currentNode = currentNode.next
	}
	return result
}
