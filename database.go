package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type logrecord struct {
	Count int
	Title string
}

// Logdatabase - main database structure
type Logdatabase struct {
	lock sync.Mutex
	db   map[string]logrecord
}

// MakeLogDatabase Create a new logdatabase
func MakeLogDatabase() *Logdatabase {
	r := &Logdatabase{lock: sync.Mutex{},
		db: make(map[string]logrecord)}
	return r
}

// Checks that the map contains the URI
func (ldb *Logdatabase) containsURI(uri string) bool {
	ldb.lock.Lock()
	defer ldb.lock.Unlock()
	_, ok := ldb.db[uri]
	return ok
}

// Updates the database
func (ldb *Logdatabase) updateURI(uri string) logrecord {
	ldb.lock.Lock()
	defer ldb.lock.Unlock()
	i, exists := ldb.db[uri]

	if !exists {
		info := ValidateURL("https://sheep.horse" + uri)

		i.Title = info.title
	}

	i.Count = i.Count + 1
	ldb.db[uri] = i
	return i
}

// Gets a value from the database
func (ldb *Logdatabase) getURI(uri string) (logrecord, bool) {
	ldb.lock.Lock()
	defer ldb.lock.Unlock()
	i, ok := ldb.db[uri]
	return i, ok
}

// DumpDatabase - copies database to an array
func (ldb *Logdatabase) DumpDatabase() ([]byte, error) {
	ldb.lock.Lock()
	defer ldb.lock.Unlock()
	result, err := json.Marshal(ldb.db)
	return result, err
}

// DumpDatabaeToFile - dumps the database to a file
func (ldb *Logdatabase) DumpDatabaeToFile(filename string) error {
	contents, err := ldb.DumpDatabase()

	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, contents, 0666)
}

// LoadDatabase - loads from a file
func (ldb *Logdatabase) LoadDatabase(filename string) {
	file, e := ioutil.ReadFile(filename)
	if e != nil {
		log.Fatalf("Could not open database file: " + filename)
		os.Exit(1)
	}

	var data map[string]logrecord
	json.Unmarshal(file, &data)
	ldb.lock.Lock()
	defer ldb.lock.Unlock()

	ldb.db = data
}
