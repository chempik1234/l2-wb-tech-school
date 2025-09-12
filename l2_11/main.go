package main

import (
	"fmt"
	"strings"
)

var stringComparator = func(a, b string) bool { return a < b }
var runeComparator = func(a, b rune) bool { return a < b }

func hash(word string) string {
	return string(mergeSort([]rune(word), runeComparator))
}

func findAllAnagrams(words []string) map[string][]string {
	result := make(map[string][]string)
	sortedWords := make(map[string]map[string]struct{})

	// firstly we should group the words
	// get a map looking like [sortedLower]: {word1, word2, word3}
	// to let only unique keys remain we use map
	//
	// merge sort of a string is m log m
	for _, word := range words {
		word = strings.ToLower(word)

		key := hash(word)
		if _, ok := sortedWords[key]; !ok {
			sortedWords[key] = make(map[string]struct{}, 1)
		}
		sortedWords[key][word] = struct{}{}
	}

	// secondly we get all sets, sort and form result like [word1]: {word1, word2, word3}
	//
	// sort is m log m again
	for _, v := range sortedWords {
		if len(v) > 1 {
			unsortedWords := make([]string, 0, len(v))
			for word, _ := range v {
				unsortedWords = append(unsortedWords, word)
			}
			sortedSet := mergeSort(unsortedWords, stringComparator)
			result[sortedSet[0]] = sortedSet
		}
	}
	return result
}

func main() {
	n := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	fmt.Println(n)
	fmt.Println(findAllAnagrams(n))

	/*
		Merge is MUCH faster than linked list!

		M := 50000
			a := make([]int, M)
			b := make([]int, M)
			for i := 0; i < len(a); i++ {
				a[i] = i
				b[i] = i
			}
			c := newLinkedList[int]()
			fmt.Println("sort linked list")
			for _, v := range a {
				c.insertSorted(v, func(a, b int) bool { return a < b })
			}
			fmt.Println("linked list!")
			time.Sleep(time.Second / 2)
			fmt.Println("sort merge")
			_ = mergeSort(b, func(a, b int) bool { return a < b })
			fmt.Println("merge!")
	*/

}
