package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type urirecord struct {
	count int
}

type visitlogserver struct {
	db *Logdatabase
	qm *QuizManager
}

type vistresult struct {
	CannonicalURI string
	Count         int
}

func canonicalizeURI(u string) (string, error) {
	result := u

	for i := 0; i < 3; i++ {
		unescaped, err := url.PathUnescape(result)
		if err != nil {
			return u, err
		}
		if unescaped == result {
			break
		}
		result = unescaped
	}

	return result, nil
}

// Handle a hit on a URL
func (s *visitlogserver) handleHit() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// parse out the cannonical URI
		uri := r.FormValue("uri")

		// sanity check
		if uri == "" {
			http.Error(w, "No Uri", http.StatusBadRequest)
			log.Print("BAD REQUEST - No uri")
			return
		}
		realURI, err := canonicalizeURI(uri)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Print("BAD REQUEST - bad uri - " + uri)
			return
		}

		var record logrecord
		if r.Method == "POST" {
			record = s.db.updateURI(realURI, "hit")
		} else {
			record, _ = s.db.getURI(realURI, "hit")
		}

		log.Print(fmt.Sprintf("%s %d", realURI, record.Count))

		result := vistresult{CannonicalURI: realURI,
			Count: record.Count}
		b, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

func (s *visitlogserver) handleStats() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "What?", http.StatusBadRequest)
			return
		}

		result, _ := s.db.DumpDatabaseRealm("hit")

		w.Header().Set("Content-Type", "application/json")
		w.Write(result)

	}
}

func (s *visitlogserver) handleAnswer() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "What? Require POST", http.StatusBadRequest)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(w, "What? Bad content type", http.StatusBadRequest)
			return
		}

		type AnswerRequest struct {
			QuizID   string
			Question int
			AnswerID string
		}

		var ar AnswerRequest

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&ar)
		if err != nil {
			http.Error(w, "What? Body did not decode", http.StatusBadRequest)
			return
		}

		if (ar.QuizID == "") || (ar.AnswerID == "") {
			http.Error(w, "What? Body did not contain the required attributes", http.StatusBadRequest)
			return
		}

		result, err := s.qm.SubmitAnswer(ar.QuizID, ar.Question, ar.AnswerID)
		if err != nil {
			log.Printf("QUIZ ERROR %s", err.Error())
			http.Error(w, "What?", http.StatusInternalServerError)
			return
		}

		jsonResult, _ := json.MarshalIndent(result, "", "  ")

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResult)

	}
}

func main() {

	f, err := os.OpenFile("/var/log/visitlog", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)
	defer f.Close()

	vs := &visitlogserver{db: MakeLogDatabase(), qm: MakeQuizManager(QuizDatabaseDirectory{path: "quizes"})}
	vs.db.LoadDatabase("visitlogdb")

	http.HandleFunc("/log", vs.handleHit())
	http.HandleFunc("/stats", vs.handleStats())
	http.HandleFunc("/quiz/submit", vs.handleAnswer())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
