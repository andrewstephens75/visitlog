package main

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func getTestQuizes() *QuizManager {
	qm := MakeQuizManager(QuizDatabaseDirectory{path: "."})

	q := question{Answers: make(map[string]answer)}
	q.Answers["fish"] = answer{Count: 5}
	q.Answers["fowl"] = answer{Count: 27}

	testQuiz := quiz{dirty: true,
		Questions: nil}

	testQuiz.Questions = append(testQuiz.Questions, q)

	qm.quizes["quiz1"] = testQuiz

	return qm
}

func TestSubmitingAnswers(t *testing.T) {

	qm := getTestQuizes()

	var tests = []struct {
		q      string
		num    int
		ans    string
		output string
		err    bool
	}{
		{"quiz1", 0, "fish", "fish {6}fowl {27}", false},
		{"dfgdgf", 0, "fish", "<nil>", true},
		{"quiz1", 27, "fish", "<nil>", true},
		{"quiz1", 0, "apple", "<nil>", true}}

	for index, test := range tests {
		result, err := qm.SubmitAnswer(test.q, test.num, test.ans)

		if (err != nil) && (test.err == false) {
			t.Errorf("Test %d had an eror", index)
			continue
		}
		if (test.err == true) && (err == nil) {
			t.Errorf("Test %d expected error", index)
			continue
		}

		var outputString = "<nil>"
		if result != nil {
			var keys []string
			for k := range result.Answers {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			var builder strings.Builder

			for _, v := range keys {
				fmt.Fprintf(&builder, "%v %v", v, result.Answers[v])
			}

			outputString = builder.String()
		}

		if test.output != outputString {
			t.Errorf("Test %d expected %q got %q", index, test.output, outputString)
		}
	}
}

func TestGetQuizIDs(t *testing.T) {
	qm := getTestQuizes()

	ids := qm.GetQuizeIDs()
	if len(ids) != 1 {
		t.Errorf("Could not get quizes")
		return
	}

	if ids[0] != "quiz1" {
		t.Errorf("Incorrect quiz id")
		return
	}
}

func TestGetQuizResults(t *testing.T) {
	qm := getTestQuizes()

	var tests = []struct {
		id          string
		output      string
		errExpected bool
	}{{"quiz1", "[{map[fish:{5} fowl:{27}]}]", false},
		{"quizdgdg1", "[]", true}}

	for index, test := range tests {
		result, err := qm.GetQuizResults(test.id)

		ok := err != nil
		if test.errExpected != ok {
			t.Errorf("Wrong error on %d", index)
			continue
		}

		if test.output != fmt.Sprintf("%v", result) {
			t.Errorf("Expected %s got %v", test.output, result)
			continue
		}

	}
}

func TestSavingAndLoading(t *testing.T) {
	qm := getTestQuizes()

	qm.saveDirtyQuizes()

	qm.quizes = make(map[string]quiz)

	_, err := qm.SubmitAnswer("quiz1", 0, "fish")
	if err != nil {
		t.Errorf("Got error " + err.Error())
	}
}

func TestIllegalFiles(t *testing.T) {
	db := QuizDatabaseDirectory{path: "validpath"}

	var tests = []struct {
		name         string
		fullPathname string
		hasError     bool
	}{{"quiz", "validpath/quiz.json", false},
		{"back\\slash", "", true},
		{"pipe|pipe", "", true}}

	for _, test := range tests {
		name, err := db.getFullPathname(test.name)
		if err != nil && test.hasError == false {
			t.Errorf("Expected no error with %q", test.name)
		} else if err == nil && test.hasError == true {
			t.Errorf("Expected an error with %q", test.name)
		} else if name != test.fullPathname {
			t.Errorf("Expect %q == %q", test.name, name)
		}
	}
}
