package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

func printLine(printNumbers bool, filename string, lineno int, lineIsFound bool, value string) {
	if len(filename) > 0 {
		if lineIsFound {
			fmt.Print(filename, ":")
		} else {
			fmt.Print(filename, "-")
		}
	}

	if printNumbers {
		if lineIsFound {
			fmt.Printf("%d:", lineno)
		} else {
			fmt.Printf("%d-", lineno)
		}
	}

	fmt.Print(value, "\n")
}

func main() {
	var err error

	aFlag := flag.Int("A", 0, "print next A strings after each found string as a context")
	bFlag := flag.Int("B", 0, "print prev B strings before ach found string as a context")
	cFlag := flag.Int("C", 0, "print C strings after and C strings before each found string as a context (overrides A and B)")
	c2Flag := flag.Bool("c", false, "print only amount of lines (a number)")
	iFlag := flag.Bool("i", false, "ignore case")
	vFlag := flag.Bool("v", false, "negate the template")
	fFlag := flag.Bool("F", false, "use template as fixed string, not a regex")
	nFlag := flag.Bool("n", false, "print numbers of lines: 1, 2, 3 etc.")

	flag.Parse()

	addLinesAfterAmount := *aFlag
	addLinesBeforeAmount := *bFlag
	printOnlyLinesAmount := *c2Flag
	ignoreCase := *iFlag
	negateTemplate := *vFlag
	useAsFixedString := *fFlag
	printNumbers := *nFlag

	// override A and B with C
	addLinesAroundAmount := *cFlag
	if addLinesAroundAmount > 0 {
		addLinesBeforeAmount = addLinesAroundAmount
		addLinesAfterAmount = addLinesAroundAmount
	}

	// если шаблон передать не после флагов, то они не считаются, потому что пакет сделали рил гении
	args := flag.Args()

	if len(args) == 0 {
		log.Fatal(errors.New("no template provided"))
	}

	templateString := args[0]
	if ignoreCase {
		templateString = strings.ToLower(templateString)
	}

	// rest args are files
	// so if only 1 arg then readers=[os.Stdin] len = 1
	//    if more args then readers=[file1, file2] len=2
	// len readers = len args
	readers := make([]*bufio.Reader, 0, len(args))
	fileNames := make([]string, 0, len(args))
	if len(args) == 1 {
		readers = append(readers, bufio.NewReader(os.Stdin))
		fileNames = append(fileNames, "")
	} else {
		// args = template, file1.txt, file2.txt
		// readers = Reader(file1.txt), Reader(file2.txt)
		var file *os.File
		for i := 1; i < len(args); i++ {
			filename := args[i]
			file, err = os.Open(filename)
			if err != nil {
				log.Fatal(fmt.Errorf("error opening file %s: %w", filename, err))
			}
			readers = append(readers, bufio.NewReader(file))
			fileNames = append(fileNames, filename)
		}
	}

	// store pointer to not occupy too much memory if unused
	var templateRegex *regexp.Regexp

	if !useAsFixedString {
		templateRegex, err = regexp.Compile(templateString)
		if err != nil {
			log.Fatal(fmt.Errorf("invalid regex template: %w", err))
			return // IDE helper to ensure templateRegex != nil
		}
	}

	readerIndex := 0
	reader := readers[readerIndex]

	var input, interpretedString string

	var counter = 0
	var lineno = 0

	// linesToPrintAfter = addLinesAfterAmount on each found line
	var linesToPrintAfter = 0
	// linesSinceLastPrint - print addLinesBeforeAmount lines up to the last printed
	var linesSinceLastPrint = 0

	// since we're reading and printing all lines in 1 iteration
	// and the file might be large
	// a FIFO buffer is used to print prev lines if a "line to print" was finally found
	// otherwise the first element is deleted from the buffer to not occupy space
	//
	// template: 4, cap: 2
	//  1 []
	//  2 [1]
	//  3 [1, 2]
	//  4 [2, 3] -> print 2, 3, 4
	//  5 []
	//  6 [5]
	// etc.
	prevLines := NewQueue(addLinesBeforeAmount)

	// defines if current line satisfies the pattern, or it's just "context" (-C, -A, -B)
	// it affects on:
	// -- whether "prev lines" are printed (print only for true value)
	// -- whether "prev lines" are cleared (same as prev.)
	// -- whether we add cur. line to "prev lines" (only for false)
	// -- whether we use delimiter of : or - (: for true, - for false)
	var currentLineIsFound = false

	for {
		currentLineIsFound = false

		if printNumbers {
			lineno++
		}

		input, err = reader.ReadString('\n')

		// При конце файла переходим к следующему, иначе показываем ошибку чтения
		// Если файлы кончились, выходим
		errors.Is(err, nil)
		if errors.Is(err, io.EOF) {
			// reset on each file
			linesToPrintAfter = 0

			readerIndex++
			if readerIndex == len(readers) {
				break
			}

			reader = readers[readerIndex]
			continue
		} else if err != nil {
			log.Fatal(err)
		}

		// убираем \n который остаётся после ReadString

		// input - изначальная версия строки
		input = strings.TrimSuffix(input, "\n")

		// interpretedString - is the version that is checked with regex
		if ignoreCase {
			interpretedString = strings.ToLower(interpretedString)
		} else {
			interpretedString = input
		}

		if negateTemplate != ((useAsFixedString && strings.Contains(interpretedString, templateString)) ||
			(!useAsFixedString && templateRegex.MatchString(interpretedString))) {
			linesToPrintAfter = addLinesAfterAmount + 1 // + 1 current line
			currentLineIsFound = true
		}

		if linesToPrintAfter > 0 {
			if printOnlyLinesAmount && currentLineIsFound {
				counter++
			} else {
				// print prev lines for context
				if currentLineIsFound {
					i := lineno - prevLines.Len()
					for line := range prevLines.All() {
						printLine(printNumbers, fileNames[readerIndex], i, false, line)
						i++
					}
				}
				printLine(printNumbers, fileNames[readerIndex], lineno, currentLineIsFound, interpretedString)
			}
			prevLines.Clear()
			linesToPrintAfter--
		} else {
			linesSinceLastPrint++
		}

		if !currentLineIsFound && addLinesBeforeAmount > 0 {
			prevLines.Append(input)
		}
	}

	if printOnlyLinesAmount {
		fmt.Println(counter)
	}
}

/*
PS C:\Users\Danis\wbtech\l2\l2_12> go run ./... -n -A 1 "fmt.(.+)" main.go go.mod
main.go:25:                     fmt.Print(filename, ":")
main.go-26-             } else {
main.go:27:                     fmt.Print(filename, "-")
main.go-28-             }
main.go:33:                     fmt.Printf("%d:", lineno)
main.go-34-             } else {
main.go:35:                     fmt.Printf("%d-", lineno)
main.go-36-             }
main.go:39:     fmt.Print(value, "\n")
main.go-40-}
main.go:100:                            log.Fatal(fmt.Errorf("error opening file %s: %w", filename, err))
main.go-101-                    }
main.go:113:                    log.Fatal(fmt.Errorf("invalid regex template: %w", err))
main.go-114-                    return // IDE helper to ensure templateRegex != nil
main.go:225:            fmt.Println(counter)
main.go-226-    }
main.go:240:main.go-25-                     fmt.Print(filename, ":")
main.go-241-main.go-26-             } else {
main.go:246:main.go-33-                     fmt.Printf("%d:", lineno)
main.go-247-main.go-34-             } else {
main.go:277:main.go-100-                            log.Fatal(fmt.Errorf("error opening file %s: %w", filename, err))
main.go-278-main.go-101-                    }
main.go:289:main.go-113-                    log.Fatal(fmt.Errorf("invalid regex template: %w", err))
main.go-290-main.go-114-                    return // IDE helper to ensure templateRegex != nil
main.go:345:main.go-206-            fmt.Println(counter)
main.go-346-main.go-207-    }


PS C:\Users\Danis\wbtech\l2\l2_12> go run ./... -n -F -C 2 if main.go go.mod
main.go-14-
main.go-15-func maxInt(a, b int) int {
main.go:16:     if a > b {
main.go-17-             return a
main.go-18-     }
main.go-21-
main.go-22-func printLine(printNumbers bool, filename string, lineno int, lineIsFound bool, value string) {
main.go:23:     if len(filename) > 0 {
main.go:24:             if lineIsFound {
main.go-25-                     fmt.Print(filename, ":")
main.go-26-             } else {
main.go-29-     }
main.go-30-
main.go:31:     if printNumbers {
main.go:32:             if lineIsFound {
main.go-33-                     fmt.Printf("%d:", lineno)
main.go-34-             } else {
main.go-64-     // override A and B with C
main.go-65-     addLinesAroundAmount := *cFlag
main.go:66:     if addLinesAroundAmount > 0 {
main.go-67-             addLinesBeforeAmount = addLinesAroundAmount
main.go-68-             addLinesAfterAmount = addLinesAroundAmount
main.go-72-     args := flag.Args()
main.go-73-
main.go:74:     if len(args) == 0 {
main.go-75-             log.Fatal(errors.New("no template provided"))
main.go-76-     }
main.go-77-
main.go-78-     templateString := args[0]
main.go:79:     if ignoreCase {
main.go-80-             templateString = strings.ToLower(templateString)
main.go-81-     }
main.go-82-
main.go-83-     // rest args are files
main.go:84:     // so if only 1 arg then readers=[os.Stdin] len = 1
main.go:85:     //    if more args then readers=[file1, file2] len=2
main.go-86-     // len readers = len args
main.go-87-     readers := make([]*bufio.Reader, 0, len(args))
main.go-87-     readers := make([]*bufio.Reader, 0, len(args))
main.go-88-     fileNames := make([]string, 0, len(args))
main.go:89:     if len(args) == 1 {
main.go-90-             readers[0] = bufio.NewReader(os.Stdin)
main.go-91-             fileNames[0] = ""
main.go-97-                     filename := args[i]
main.go-98-                     file, err = os.Open(filename)
main.go:99:                     if err != nil {
main.go-100-                            log.Fatal(fmt.Errorf("error opening file %s: %w", filename, err))
main.go-101-                    }
main.go-105-    }
main.go-106-
main.go:107:    // store pointer to not occupy too much memory if unused
main.go-108-    var templateRegex *regexp.Regexp
main.go-109-
main.go-109-
main.go:110:    if !useAsFixedString {
main.go-111-            templateRegex, err = regexp.Compile(templateString)
main.go-111-            templateRegex, err = regexp.Compile(templateString)
main.go:112:            if err != nil {
main.go-113-                    log.Fatal(fmt.Errorf("invalid regex template: %w", err))
main.go-114-                    return // IDE helper to ensure templateRegex != nil
main.go-130-    prevLines := NewQueue(addLinesBeforeAmount)
main.go-131-
main.go:132:    // defines if current line satisfies the pattern, or it's just "context" (-C, -A, -B)
main.go-133-    var currentLineIsFound = false
main.go-134-
main.go-136-            currentLineIsFound = false
main.go-137-
main.go:138:            if printNumbers {
main.go-139-                    lineno++
main.go-140-            }
main.go-145-            // Если файлы кончились, выходим
main.go-146-            errors.Is(err, nil)
main.go:147:            if errors.Is(err, io.EOF) {
main.go-148-                    // reset on each file
main.go-149-                    linesToPrintAfter = 0
main.go-150-
main.go-151-                    readerIndex++
main.go:152:                    if readerIndex == len(readers) {
main.go-153-                            break
main.go-154-                    }
main.go-156-                    reader = readers[readerIndex]
main.go-157-                    continue
main.go:158:            } else if err != nil {
main.go-159-                    log.Fatal(err)
main.go-160-            }
main.go-166-
main.go-167-            // interpretedString - is the version that is checked with regex
main.go:168:            if ignoreCase {
main.go-169-                    interpretedString = strings.ToLower(interpretedString)
main.go-170-            } else {
main.go-172-            }
main.go-173-
main.go:174:            if negateTemplate != ((useAsFixedString && strings.Contains(interpretedString, templateString)) ||
main.go-175-                    (!useAsFixedString && templateRegex.MatchString(interpretedString))) {
main.go-176-                    linesToPrintAfter = addLinesAfterAmount + 1 // + 1 current line
main.go-178-            }
main.go-179-
main.go:180:            if linesToPrintAfter > 0 {
main.go:181:                    if printOnlyLinesAmount && currentLineIsFound {
main.go-182-                            counter++
main.go-183-                    } else {
main.go-183-                    } else {
main.go-184-                            // print prev lines for context
main.go:185:                            if currentLineIsFound {
main.go-186-                                    i := lineno - prevLines.Len()
main.go-187-                                    for line := range prevLines.All() {
main.go-198-            }
main.go-199-
main.go:200:            if !currentLineIsFound && addLinesBeforeAmount > 0 {
main.go-201-                    prevLines.Append(input)
main.go-202-            }
main.go-203-    }
main.go-204-
main.go:205:    if printOnlyLinesAmount {
main.go-206-            fmt.Println(counter)
main.go-207-    }
*/
