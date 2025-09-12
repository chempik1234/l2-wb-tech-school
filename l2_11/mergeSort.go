package main

import "sync"

// mergeSort accepts a slice and comparator “func(a, b T) bool“ than returns true if a < b
//
// example: func (a, b rune) bool { return a < b }
//
// classic algorithm, “m log m“ complexity
func mergeSort[T any](items []T, comparator func(a, b T) bool) []T {
	if len(items) < 2 {
		return items
	}
	border := len(items) / 2

	wg := sync.WaitGroup{}
	wg.Add(2)

	var half1 []T
	var half2 []T

	go func() {
		defer wg.Done()
		half1 = mergeSort(items[:border], comparator)
	}()

	go func() {
		defer wg.Done()
		half2 = mergeSort(items[border:], comparator)
	}()

	wg.Wait()

	return merge(
		half1,
		half2,
		comparator,
	)
}

// merge merges 2 slices using the comparator func (check mergeSort)
func merge[T any](a, b []T, comparator func(a, b T) bool) []T {
	i, j := 0, 0
	result := make([]T, 0, len(a)+len(b))
	for i < len(a) && j < len(b) {
		if comparator(a[i], b[j]) {
			result = append(result, a[i])
			i++
		} else {
			result = append(result, b[j])
			j++
		}
	}
	// remaining part of either A or B contains elements that are surely greater than ones already in result

	// so only 1 of 2 happens
	for ; i < len(a); i++ {
		result = append(result, a[i])
	}
	for ; j < len(b); j++ {
		result = append(result, b[j])
	}
	return result
}
