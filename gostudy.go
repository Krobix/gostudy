package main

import (
	"fmt"
	"os"
	"bufio"
	"strconv"
	"log"
	"strings"
	"encoding/gob"
	"bytes"
	"math/rand"
	"time"
)

/*DebugMode : if debug messages should be displayed */
const DebugMode bool = false

//Question is a question :|
type Question struct {
	QuestionText string
	Answer string
}

//Assignment is a full quiz (group of Questions, pluss additional config data)
type Assignment struct {
	Questions []Question
	ShowAnswersAmount int
	QuestionsAmount int
}

func debug(txt string) {
	if DebugMode {
		fmt.Printf("[DEBUG]%s \n", txt)
	}
}

func createQuestion(QuestionText string, Answer string) *Question {
	tmpq := new(Question)
	tmpq.QuestionText = QuestionText
	tmpq.Answer = Answer
	return tmpq
}

func createAssignment() *Assignment{
	return new(Assignment)
}

func encodeAssignment(a *Assignment) bytes.Buffer{
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(*a)
	if err != nil {
		log.Fatalln(err)
	}
	return b
}

func decodeAssignment(b []byte) *Assignment {
	a := Assignment{}
	bb := bytes.Buffer{}
	bb.Write(b)
	d := gob.NewDecoder(&bb)
	err := d.Decode(&a)
	if err != nil {
		log.Fatalln(err)
	}
	return &a
}

func decodeAssignmentFromFile(fname string) *Assignment{
	f, err := os.Open(fname)
	debug("Opened file and created empty bytes")
	if err != nil {
		log.Fatalln(err)
	}
	finfo, ferr := f.Stat()
	if ferr != nil {
		log.Fatalln(ferr)
	}
	fsize := finfo.Size()
	b := make([]byte, fsize)
	_, err = f.Read(b)
	debug("Attempting to read file")
	if err != nil {
		log.Fatalln(err)
	}
	return decodeAssignment(b)
}

func chooseShownAnswers(assignment *Assignment, correct *string) map[int]*string {
	m := make(map[int]*string)
	for i := 0; i < assignment.ShowAnswersAmount; i++ {
		choice := rand.Intn(assignment.QuestionsAmount-1)
		m[i] = &(assignment.Questions[choice].Answer)
		if *(m[i]) == *correct {
			i--
		}
	}
	choiceForCorrect := rand.Intn(assignment.ShowAnswersAmount)
	m[choiceForCorrect] = correct
	return m
}

func askQuestion(question *Question, assignment *Assignment, reader *bufio.Reader) bool {
	fmt.Printf("Question: %s\n", question.QuestionText)
	correctAnswer := &question.Answer
	shownAnswers := chooseShownAnswers(assignment, correctAnswer)
	for i := 0; i < assignment.ShowAnswersAmount; i++ {
		fmt.Printf("[%d]%s\n", i, *(shownAnswers[i]))
	}
	fmt.Println("(Enter the number of the answer you believe to be correct).")
	answer, _ := reader.ReadString('\n')
	answer = strings.Trim(answer, "\n")
	intAnswer, err := strconv.Atoi(answer)
	if err != nil {
		log.Fatal(err)
	}
	return shownAnswers[intAnswer] == correctAnswer
}

func createAssignmentInteractive(reader *bufio.Reader) {
	assignment := createAssignment()
	fmt.Println("Specify the total number of questions.")
	assignment.ShowAnswersAmount = 4
	text, _ := reader.ReadString('\n')
	text = strings.Trim(text, "\n")
	QuestionsAmount, err := strconv.Atoi(text)
	if err != nil {
		log.Fatalln(err)
	}
	assignment.QuestionsAmount = QuestionsAmount
	debug("QuestionAmount assigned")
	for i := 0; i < assignment.QuestionsAmount; i++ {
		debug("for loop started in createAssignmentInteractive")
		fmt.Println("Enter the question.")
		questext, _ := reader.ReadString('\n')
		fmt.Println("Enter the correct answer.")
		answer, _ := reader.ReadString('\n')
		qobj := createQuestion(questext, answer)
		assignment.Questions = append(assignment.Questions, *qobj)
	}
	encodedAssignment := encodeAssignment(assignment)
	f, err := os.Create("assignment.bin")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	_, ferr := encodedAssignment.WriteTo(f)
	if ferr != nil {
		log.Fatalln(ferr)
	}
	fmt.Println("Created studyset written to assignment.bin.")
	os.Exit(0)
}

func studyInteractive(a *Assignment, reader *bufio.Reader) {
	score := 0
	rand.Shuffle(a.QuestionsAmount, func(i, j int){a.Questions[i], a.Questions[j] = a.Questions[j], a.Questions[i]})
	for i := 0; i < a.QuestionsAmount; i++ {
		correct := askQuestion(&(a.Questions[i]), a, reader)
		if correct {
			fmt.Println("Correct")
			score++
		} else {
			fmt.Println("Incorrect")
		}
	}
	fmt.Printf("You finished with a %d/%d\n",score,a.QuestionsAmount)
}

func main(){
	gob.Register(Question{})
	gob.Register(Assignment{})
	rand.Seed(time.Now().UnixNano())
	argv := os.Args[1:]
	reader := bufio.NewReader(os.Stdin)
	if len(argv) == 0 {
		fmt.Println("You must enter an argument: create to create a new studyset, or quiz <filename> to study an existing one.")
	} else if argv[0] == "create" {
		createAssignmentInteractive(reader)
	} else if argv[0] == "debugread" {
		if len(argv) == 1 {
			fmt.Println("A filename is required")
			os.Exit(1)
		}
		assignment := decodeAssignmentFromFile(argv[1])
		for i := 0; i < assignment.QuestionsAmount; i++ {
			fmt.Println("QUESTION:")
			fmt.Println(assignment.Questions[i].QuestionText)
			fmt.Println("ANSWER:")
			fmt.Println(assignment.Questions[i].Answer)
		}
	} else if argv[0] == "quiz" {
		if len(argv) == 1 {
			fmt.Println("A filename is required")
			os.Exit(1)
		}
		assignment := decodeAssignmentFromFile(argv[1])
		studyInteractive(assignment, reader)
	}
}