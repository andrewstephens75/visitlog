package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

type logrecord struct {
	Count int
	Title string
}

// Logdatabase - main database structure
// database is a map of maps
// realm -> uri -> logrecord
type Logdatabase struct {
	lock     *sync.Mutex
	db       map[string]map[string]logrecord
	dirty    bool
	sweeper  *time.Ticker
	filename string
}

// MakeLogDatabase Create a new logdatabase
func MakeLogDatabase() *Logdatabase {
	r := Logdatabase{lock: &sync.Mutex{},
		db:    make(map[string]map[string]logrecord),
		dirty: false}

	r.sweeper = time.NewTicker(10 * time.Second)
	go func() {
		handledError := false
		for {
			<-r.sweeper.C
			err := r.DumpDatabaseToFile("")
			if err != nil {
				if handledError == false {
					log.Printf("Error when saving database %s", err.Error())
					handledError = true
				}
			} else {
				handledError = true
			}

		}
	}()
	return &r
}

// Checks that the map contains the URI
func (ldb *Logdatabase) containsURI(uri string, realm string) bool {
	ldb.lock.Lock()
	defer ldb.lock.Unlock()

	realmMap, ok := ldb.db[realm]
	if ok {
		_, ok = realmMap[uri]
	}
	return ok
}

// Updates the database
func (ldb *Logdatabase) updateURI(uri string, realm string) logrecord {
	ldb.lock.Lock()
	defer ldb.lock.Unlock()
	realmMap, exists := ldb.db[realm]

	if !exists {
		return logrecord{Title: "", Count: 0}
	}

	i, exists := realmMap[uri]

	if !exists {
		// special case handling for the "hit" realm
		if realm == "hit" {
			info := ValidateURL("https://sheep.horse" + uri)

			if info.err != nil {
				return logrecord{Title: "", Count: 0}
			}
			i.Title = info.title
		} else {
			return logrecord{Title: "", Count: 0}
		}
	}

	i.Count = i.Count + 1
	realmMap[uri] = i
	ldb.dirty = true
	return i
}

// Gets a value from the database
func (ldb *Logdatabase) getURI(uri string, realm string) (logrecord, bool) {
	ldb.lock.Lock()
	defer ldb.lock.Unlock()

	var i logrecord
	realmdb, ok := ldb.db[realm]
	if !ok {
		return i, ok
	}
	i, ok = realmdb[uri]
	return i, ok
}

func (ldb *Logdatabase) marshalDatabase() ([]byte, error) {
	result, err := json.MarshalIndent(ldb.db, "", "  ")
	return result, err
}

// DumpDatabaseRealm - copies part of the database to an array
func (ldb *Logdatabase) DumpDatabaseRealm(realm string) ([]byte, error) {
	ldb.lock.Lock()
	defer ldb.lock.Unlock()

	realmdb, found := ldb.db[realm]
	if !found {
		return nil, errors.New("Realm not found")
	}

	result, err := json.MarshalIndent(realmdb, "", "  ")

	return result, err
}

// DumpDatabaseToFile - dumps the database to a file
func (ldb *Logdatabase) DumpDatabaseToFile(filename string) error {
	ldb.lock.Lock()
	defer ldb.lock.Unlock()

	if filename == "" {
		filename = ldb.filename
	}
	if ldb.dirty == false {
		return nil
	}
	contents, err := ldb.marshalDatabase()

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, contents, 0666)
	if err == nil {
		ldb.dirty = false
	}
	return err
}

// LoadDatabase - loads from a file
func (ldb *Logdatabase) LoadDatabase(filename string) {
	file, e := ioutil.ReadFile(filename)
	if e != nil {
		log.Fatalf("Could not open database file: " + filename)
		os.Exit(1)
	}

	var data map[string]map[string]logrecord
	json.Unmarshal(file, &data)
	ldb.lock.Lock()
	defer ldb.lock.Unlock()

	ldb.db = data
	ldb.dirty = false
	ldb.filename = filename
}
