package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"sync"
	"time"
)

type answer struct {
	Count int
}

type question struct {
	Answers map[string]answer
}

type quiz struct {
	dirty     bool
	Questions []question
}

type cmdUpdateRequest struct {
	quizID   string
	question int
	answerID string
}

type cmdUpdateResponse struct {
	e       error
	answers map[string]answer
}

// QuizDatabase - exposes an interface for loading and saving quizes
type QuizDatabase interface {
	LoadQuiz(name string) (quiz, error)
	SaveQuiz(q *quiz, name string) error
}

type QuizDatabaseDirectory struct {
	path string
}

func (qdb QuizDatabaseDirectory) getFullPathname(name string) (string, error) {
	const illegals = "./\\%$!@| \"<>?"

	for _, r := range illegals {
		if strings.ContainsRune(name, r) {
			return "", fmt.Errorf("%q contains illegal characters", name)
		}
	}

	return (path.Join(qdb.path, name) + ".json"), nil
}

// LoadQuiz - returns a quiz
func (qdb QuizDatabaseDirectory) LoadQuiz(name string) (quiz, error) {
	fullname, e := qdb.getFullPathname(name)
	if e != nil {
		return quiz{}, e
	}

	file, e := ioutil.ReadFile(fullname)
	if e != nil {
		return quiz{}, e
	}

	var result quiz
	err := json.Unmarshal(file, &result)
	if err != nil {
		return quiz{}, err
	}

	result.dirty = false
	return result, nil
}

// SaveQuiz - saves quiz to a file
func (qdb QuizDatabaseDirectory) SaveQuiz(q *quiz, name string) error {
	fullname, e := qdb.getFullPathname(name)
	if e != nil {
		return e
	}

	buffer, err := json.MarshalIndent(q, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fullname, buffer, 0666)
	return err
}

// QuizManager Managers the quizes and exposes the results via channels
type QuizManager struct {
	lock     *sync.Mutex
	database QuizDatabase
	quizes   map[string]*quiz
	sweeper  *time.Ticker
}

// MakeQuizManager - create a quizmanager
func MakeQuizManager(db QuizDatabase) *QuizManager {
	r := QuizManager{lock: &sync.Mutex{},
		database: db,
		quizes:   make(map[string]*quiz),
		sweeper:  time.NewTicker(10 * time.Second)}

	go func() {
		for {
			<-r.sweeper.C
			r.saveDirtyQuizes()
		}
	}()

	return &r
}

func (qm QuizManager) shutdown() {
	qm.sweeper.Stop()
	qm.saveDirtyQuizes()
	qm.database = nil
	qm.quizes = nil
}

func (qm QuizManager) getQuiz(quizID string) (*quiz, error) {
	var result *quiz
	result, ok := qm.quizes[quizID]
	if !ok {
		result, err := qm.database.LoadQuiz(quizID)
		if err != nil {
			return nil, err
		}

		qm.quizes[quizID] = &result
		return qm.quizes[quizID], nil
	}
	return result, nil
}

func (qm QuizManager) saveDirtyQuizes() {
	qm.lock.Lock()
	defer qm.lock.Unlock()
	fmt.Printf("Saving Dirty Datbases\n")
	for k, v := range qm.quizes {
		if v.dirty == true {
			err := qm.database.SaveQuiz(v, k)

			if err != nil {
				log.Printf("Could not save %q - %s", k, err.Error())
			}
			v.dirty = false
		}
	}
}

// SubmitAnswer - submit an answer, return all the answers to the same question
func (qm *QuizManager) SubmitAnswer(quizID string, questionNum int, answerID string) (*question, error) {
	qm.lock.Lock()
	defer qm.lock.Unlock()
	quiz, err := qm.getQuiz(quizID)
	if err != nil {
		return nil, fmt.Errorf("quizID %q not found", quizID)
	}

	if (questionNum < 0) || (questionNum >= len(quiz.Questions)) {
		return nil, fmt.Errorf("question %d out of bounds", questionNum)
	}

	q := quiz.Questions[questionNum]

	foundAnswer, ok := q.Answers[answerID]
	if !ok {
		return nil, fmt.Errorf("answerID %q not found", answerID)
	}

	foundAnswer.Count = foundAnswer.Count + 1
	quiz.dirty = true
	q.Answers[answerID] = foundAnswer

	log.Printf("QUIZ %q[%d] = %q", quizID, questionNum, answerID)

	answersCopy := make(map[string]answer)
	for k, v := range q.Answers {
		answersCopy[k] = v
	}
	// copy the question
	questionCopy := question{
		Answers: answersCopy}

	return &questionCopy, nil
}

// GetQuizeIDs - returns the IDs of all quizes
func (qm QuizManager) GetQuizeIDs() []string {
	qm.lock.Lock()
	defer qm.lock.Unlock()
	var r []string
	for k := range qm.quizes {
		r = append(r, k)
	}
	return r
}

// GetQuizResults - returns a copy of the list of questions in the quiz
func (qm QuizManager) GetQuizResults(quizID string) ([]question, error) {
	qm.lock.Lock()
	defer qm.lock.Unlock()
	quiz, err := qm.getQuiz(quizID)
	if err != nil {
		return nil, fmt.Errorf("quizID %q not found", quizID)
	}

	var result []question
	for _, q := range quiz.Questions {
		newQuestion := question{Answers: make(map[string]answer)}
		for k, v := range q.Answers {
			newQuestion.Answers[k] = v
		}
		result = append(result, newQuestion)
	}
	return result, nil
}
