package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	//"runtime"
)

const countPosibleAnswers = 4

type myScaner struct {
	sync.Mutex
	s string
}

type Quiz struct {
	sync.Mutex
	Timer             *time.Timer
	CustomInputTime   time.Duration
	QuestionAndAnswer map[string]string
	TotalQuestions    int
	QuestionsAnswered int
	Marks             int
	Right             int
	Wrong             int
	MaxQuestions      int
}

func (s *myScaner) myScan() string {
	var input string
	//s.Lock()
	reader := bufio.NewReader(os.Stdin)
	input, _ = reader.ReadString('\n')
	//fmt.Scan(&input)
	input = strings.TrimSpace(strings.ToLower(input))
	s.s = input
	//s.Unlock()
	return s.s
}

func myQuit(s *myScaner) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT) //os.Interrupt
	go func() {
		for {
			select {
			case <-c:
				s.Lock()

				var choice string
				fmt.Printf("\nReally quit? [y/n] > ")
				//fmt.Scan(&choice)
				choice = s.myScan()
				if choice == "y" {
					cleanup()
					os.Exit(0) // definitely not the best way to exit
				}

				s.Unlock()
			}
		}
	}()
}

func cleanup() {
	fmt.Println("Bye:)")
}

func helpPrinter() {
	text := "\n\tQuizz Game\n\nDESCRIPTION\n\t..."
	fmt.Println(text)
	os.Exit(0)
}

func newQuiz() *Quiz {
	return &Quiz{
		QuestionAndAnswer: make(map[string]string, 100),
		Timer:             time.NewTimer(3 * time.Minute),
	}
}

var s myScaner

func main() {

	//	for {
	//        runtime.Gosched()
	//    }

	var q = newQuiz()
	file := flag.String("file", "questions.csv", "path for file")
	t := flag.Int("time", 30, "Give custom game duration time(In Seconds)")
	maxq := flag.Int("maxq", 3, "Count of questions")
	h := flag.Bool("h", false, "print description")
	flag.Parse()
	if *h {
		helpPrinter()
	}
	q.MaxQuestions = *maxq
	sec := fmt.Sprintf("%ds", *t)
	duration, err := time.ParseDuration(sec)
	if err != nil {
		fmt.Println(err)
	}
	q.CustomInputTime = duration
	if q.readFromCSV(*file) != nil {
		log.Fatal(err)
	}
	q.userInterface(&s)
}

func (q *Quiz) userInterface(s *myScaner) {
	var input string
	for {
		if !q.Timer.Reset(3 * time.Minute) {
			fmt.Println("Timer reset failed")
		}
		fmt.Printf("\n\t     WELCOME TO QUIZ   \n\n")
		fmt.Printf("\n")
		//fmt.Printf("\n\t\t1.Start quiz\n")
		fmt.Printf("\t\tPress 0 for Exit\n\n")
		fmt.Printf("Enter your name:")
		input = s.myScan()
		switch input {
		case "0":
			os.Exit(0)
		default:
			q.quizGame()
		}
	}
}

// readFromCSV read data from a csv file based on our input.
func (q *Quiz) readFromCSV(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	r := csv.NewReader(bufio.NewReader(file))
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var posibleAnswers string

	for i, k := range records {
		if i >= q.MaxQuestions {
			break
		}
		for j, l := range k {
			if j == 0 {
				posibleAnswers = ""
				for numAnswer := 1; countPosibleAnswers >= numAnswer; numAnswer++ {
					posibleAnswers += fmt.Sprintf("%8d) ", numAnswer) + records[i][j+numAnswer] + "\n"
				}
				q.QuestionAndAnswer[l+"\n"+posibleAnswers] = records[i][j+1+countPosibleAnswers] //last number right answer

			}
		}
	}
	q.TotalQuestions = len(q.QuestionAndAnswer)
	return nil
}

// quizGame
func (q *Quiz) quizGame() {

	//var input string
	q.Marks = 0
	q.Right = 0
	q.Wrong = 0
	q.QuestionsAnswered = 0

	if q.Timer.Reset(q.CustomInputTime) {
		fmt.Printf("\nYour time start now...You have total of %d Seconds", q.CustomInputTime/1000000000)
	}

	go func() {
		select {
		case <-q.Timer.C:
			q.Timer.Stop()
			fmt.Println("\nSorry time out, Please try again later!!!")
			fmt.Printf("\nTotal questions: %d , You Answered: %d", q.TotalQuestions, q.QuestionsAnswered)
			fmt.Printf("\nRight answers: %d, Wrong answers: %d", q.Right, q.Wrong)

			os.Exit(0)

			//default:

		}
	}()

	go myQuit(&s)

	for question, answer := range q.QuestionAndAnswer {
	k:
		fmt.Printf("\n%sWrite the right number: ", question)
		input := s.myScan()
		inp, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid number!")
			goto k
		}
		ans, err := strconv.Atoi(answer)
		if err != nil {
			log.Fatal(err)
		}
		if inp == ans {
			fmt.Println("Right answer!")
			q.Right++
			q.Marks++
			q.QuestionsAnswered++
			continue
		}
		q.Wrong++
		q.QuestionsAnswered++
		fmt.Println("Wrong answer!")
		fmt.Printf("Right answer is %s \n", answer)
	}

	fmt.Printf("\nTotal questions: %d, You Answered: %d", q.TotalQuestions, q.QuestionsAnswered)
	fmt.Printf("\nRight answers: %d, Wrong answers: %d", q.Right, q.Wrong)

}
