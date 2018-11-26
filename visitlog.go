package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
)

type urirecord struct {
	count int
}

type visitlogserver struct {
	lock sync.Mutex
	db   map[string]urirecord
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

		s.lock.Lock()
		i, _ := s.db[realURI]
		if r.Method == "POST" {
			i.count = i.count + 1
			s.db[realURI] = i
		}
		s.lock.Unlock()

		log.Print(fmt.Sprintf("%s %d", realURI, i.count))

		result := vistresult{CannonicalURI: realURI,
			Count: i.count}
		b, _ := json.Marshal(result)
		w.Write(b)
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

	vs := &visitlogserver{lock: sync.Mutex{},
		db: make(map[string]urirecord)}

	http.HandleFunc("/log", vs.handleHit())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
