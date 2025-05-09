package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"
)

func main() {
	fname, timeOut := getArgs()
	records, err := readCsv(fname)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan bool, 1)
	fmt.Printf("You've have %ds second to complete the game\n", timeOut)
	fmt.Print("Are you ready? n/Y ")

	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()
	if err != nil {
		log.Fatalf("timer failed: %v", err)
		return
	}

	if unicode.ToLower(char) != 'y' {
		fmt.Println("Game End, happy coding!")
		return
	}

	go timer(done, timeOut)

	quizzes := toQuizzes(records)
	gameLoop(quizzes, done)
}

func getArgs() (string, int) {
	var csvName string
	var timeOut int
	const defaultFname string = "Specify csv name. Default is program"
	const defaultTimeOut = 30

	flag.StringVar(&csvName, "f", defaultFname, "./quiz-ex [-f program]")
	flag.IntVar(&timeOut, "t", defaultTimeOut, "./quiz-ex [-t 30]")
	flag.Parse()

	if csvName == defaultFname {
		csvName = "program"
	}

	return csvName, timeOut
}

func readCsv(fname string) ([][]string, error) {
	f, err := os.Open("./assets/" + fname + ".csv")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return [][]string{}, fmt.Errorf("file name: %s not existed", fname+".csv")
		}
		return [][]string{}, err
	}

	defer f.Close()
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return data, nil
}

type quiz struct {
	question string
	ans      string
}

func toQuizzes(data [][]string) []quiz {
	var quizzes []quiz

	for _, r := range data {
		if len(r) < 2 {
			continue
		}
		var quiz quiz
		for i, c := range r {
			if i == 0 {
				quiz.question = c
			} else {
				quiz.ans = c
			}
		}
		quizzes = append(quizzes, quiz)
	}
	return quizzes
}

func gameLoop(quizzes []quiz, done <-chan bool) {
	fmt.Println("Welcomes to golang quizzes!")
	scanner := bufio.NewScanner(os.Stdin)
	totalScore := 0

	for i, quiz := range quizzes {
		if isDone := <-done; isDone {
			break
		}
		fmt.Printf("%d: %s\n", i+1, quiz.question)
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		uAns := scanner.Text()
		if strings.ToLower(strings.TrimSpace(uAns)) == quiz.ans {
			totalScore += 1
		}
	}

	errScan := scanner.Err()
	if errScan != nil {
		log.Fatal(errScan)
	}
	fmt.Printf("Game End: %d/%d correct\n", totalScore, len(quizzes))
}

func timer(done chan<- bool, timeOut int) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	startTime := time.Now()

	for range ticker.C {
		if int(time.Since(startTime)/time.Second) >= timeOut {
			done <- true
			return
		} else {
			done <- false
		}
	}
}
