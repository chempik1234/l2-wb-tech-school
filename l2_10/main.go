package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

func memoryUnitToInt(s string) (int64, error) {
	lastChar, _ := utf8.DecodeLastRune([]byte(s))

	var valueAsInt64 int64
	var err error

	if '0' <= lastChar && lastChar <= '9' {
		valueAsInt64, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("error parsing int '%s': %w", s, err)
		}
	} else {
		s = s[:len(s)-1]
		valueAsInt64, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("error parsing int with memory suffix '%s': %w", s, err)
		}
	}

	var multiplier int64

	switch lastChar {
	case 'B':
		multiplier = 8
	case 'K':
		multiplier = 8 * 1024
	case 'M':
		multiplier = 8 * 1024 * 1024
	case 'G':
		multiplier = 8 * 1024 * 1024 * 1024
	case 'T':
		multiplier = 8 * 1024 * 1024 * 1024 * 1024
	default:
		multiplier = 1
	}
	return multiplier * valueAsInt64, nil
}

func monthToInt(s string) (int, error) {
	switch s {
	case "Jan":
		return 1, nil
	case "Feb":
		return 2, nil
	case "Mar":
		return 3, nil
	case "Apr":
		return 4, nil
	case "May":
		return 5, nil
	case "Jun":
		return 6, nil
	case "Jul":
		return 7, nil
	case "Aug":
		return 8, nil
	case "Sep":
		return 9, nil
	case "Oct":
		return 10, nil
	case "Nov":
		return 11, nil
	case "Dec":
		return 12, nil
	default:
		return 0, errors.New("invalid month")
	}
}

type node struct {
	value string
	next  *node
}

type linkedList struct {
	head *node
}

func newLinkedList() *linkedList {
	return &linkedList{}
}

// insertSorted вставляет элемент в связный список
//
// при проходе порядок сортировки оценивается так:
// если сравнивать надо как числа, то результат сравнения = число(значение) > число(из списка)
// если строки, то сравнивать значение и из списка напрямую
//
// если сортировка по возрастанию, результат должен быть false
// если сортировка по убыванию, то результат должен быть true
//
//	найдя такой результат, выходим и вставляем
//
// compareResult == reverse -> break -> insert
//
// для сравнения в форматах памяти числа хранятся в int64, потому что битов в терабайте очень много
// для сравнения в формате числа и месяца строковые значения преобразуются в int
func (l *linkedList) insertSorted(value string, reverse bool, onlyUnique bool, asNumber bool, asMonth bool, asMemory bool) (affectedOrder bool, err error) {
	currentNode := l.head

	var valueAsInt int
	var currentAsInt int
	var valueAsInt64 int64
	var currentAsInt64 int64

	if asNumber {
		valueAsInt, err = strconv.Atoi(value)
		if err != nil {
			return false, fmt.Errorf("invalid int number to insert %s: %w", value, err)
		}
	} else if asMonth {
		valueAsInt, err = monthToInt(value)
		if err != nil {
			return false, fmt.Errorf("invalid month to insert %s: %w", value, err)
		}
	} else if asMemory {
		valueAsInt64, err = memoryUnitToInt(value)
		if err != nil {
			return false, fmt.Errorf("invalid memory unit to insert %s: %w", value, err)
		}
	}

	var compareResult bool
	var prevNode *node
	for currentNode != nil {
		if asNumber {
			currentAsInt, err = strconv.Atoi(currentNode.value)
			if err != nil {
				return false, fmt.Errorf("invalid int number in list %s: %w", currentNode.value, err)
			}
			compareResult = valueAsInt > currentAsInt
		} else if asMonth {
			currentAsInt, err = monthToInt(currentNode.value)
			if err != nil {
				return false, fmt.Errorf("invalid month to insert %s: %w", currentNode.value, err)
			}
			compareResult = valueAsInt > currentAsInt
		} else if asMemory {
			currentAsInt64, err = memoryUnitToInt(currentNode.value)
			if err != nil {
				return false, fmt.Errorf("invalid memory in list %s: %w", currentNode.value, err)
			}
			compareResult = valueAsInt64 > currentAsInt64
		} else {
			compareResult = value > currentNode.value
		}

		if compareResult == reverse {
			break
		}

		prevNode = currentNode
		currentNode = currentNode.next
		if currentNode == nil {
			break
		}
	}

	if currentNode != nil && currentNode.value == value && onlyUnique {
		return
	}

	affectedOrder = currentNode != nil

	if prevNode == nil {
		l.head = &node{value, l.head}
		return
	}
	prevNode.next = &node{value, currentNode}
	return
}

func (l *linkedList) string() string {
	result := make([]string, 0)
	currentNode := l.head
	for currentNode != nil {
		result = append(result, currentNode.value)
		currentNode = currentNode.next
	}
	return strings.Join(result, "\n")
}

func main() {
	/*
		kFlag := flag.Int("k", -1, "column to sort, count from 0, set -1 to disable")
		// converters priority from high to low
		nFlag := flag.Bool("n", false, "interpret sorted part of line as number")
		mFlag := flag.Bool("M", false, "interpret sorted part of line as month: Jan, Feb, Mar ...")
		hFlag := flag.Bool("h", false, "interpret sorted part of line as memory volume: 1M = 8192")
		//
		rFlag := flag.Bool("r", false, "reverse")
		uFlag := flag.Bool("u", false, "only unique")
		bFlag := flag.Bool("b", false, "ignore trailing blanks")
		cFlag := flag.Bool("c", false, "check and tell if data is sorted")
		flag.Parse()
	*/

	// Сканируем флаги вручную! Потому что надо обеспечить комбинации

	// region flags
	var err error

	var columnToSort = -1
	var asNumber = false
	var asMonth = false
	var asMemory = false
	var reverse = false
	var onlyUnique = false
	var trimTrailing = false
	var tellIfUnsorted = false

	scanningK := false

	for _, flagCombination := range os.Args[1:] {
		if scanningK {
			columnToSort, err = strconv.Atoi(flagCombination)
			if err != nil {
				log.Fatal("invalid -k column to sort")
			}
			scanningK = false
			continue
		}
		for _, flagRune := range flagCombination {
			switch flagRune {
			case 'k':
				scanningK = true
			case 'n':
				asNumber = true
			case 'M':
				asMonth = true
			case 'h':
				asMemory = true
			case 'r':
				reverse = true
			case 'u':
				onlyUnique = true
			case 'b':
				trimTrailing = true
			case 'c':
				tellIfUnsorted = true
			}
		}
	}
	// endregion

	// В linked list просто вставлять элемент в нужное место
	result := newLinkedList()

	affectedOrder := false

	// читаем данные из STDIN, которые нам присылает пайплайн
	reader := bufio.NewReader(os.Stdin)
	var input string
	for {
		input, err = reader.ReadString('\n')

		// При конце файла выходим, иначе показываем ошибку чтения
		errors.Is(err, nil)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		// убираем \n который остаётся после ReadString

		// input - изначальная версия строки
		input = strings.TrimSuffix(input, "\n")

		// interpretedString - обрабатываемая часть строки (колонка и/или строка обрезанная по пробелам)
		interpretedString := input
		if trimTrailing {
			interpretedString = strings.TrimRight(interpretedString, " ")
		}

		// пытаемся взять колонку в качестве interpretedString если номер указан
		if columnToSort >= 0 {
			columns := strings.Split(interpretedString, "\t")
			if columnToSort >= len(columns) {
				log.Fatal(fmt.Errorf("column too high: %d, only %d columns", columnToSort, len(columns)))
			}

			interpretedString = columns[columnToSort]
		}

		// готово, остальные преобразования сделает insert
		affectedOrder, err = result.insertSorted(interpretedString, reverse, onlyUnique, asNumber, asMonth, asMemory)
		if err != nil {
			log.Fatal("error inserting value in list:", err)
		}
	}

	// результат готов после прохода по STDIN, выведем
	fmt.Println(result.string())

	// если же при получении строк пришлось некоторые переставить, значит данные не сортированы
	// есть флаг, который предписывает вывести это
	if tellIfUnsorted && affectedOrder {
		fmt.Println("-- input data is not sorted")
	}
}
