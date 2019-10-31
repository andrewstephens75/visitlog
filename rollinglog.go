package main

import (
	"container/list"
	"time"
)

// LogEntry - a single entry plus timestamp
type LogEntry struct {
	url       string
	timestamp time.Time
}

// RollingLog - Keeps a log of strings, descarding entries older than the specified duration
type RollingLog struct {
	entries *list.List
	keepfor time.Duration
}

// MakeRollingLog - makes a empty rolling log with the specified duration
func MakeRollingLog(k time.Duration) RollingLog {
	return RollingLog{entries: list.New(),
		keepfor: k}
}

// AddEntry adds a now entry to the rolling log, perhaps cleaning up old entries
func (rl *RollingLog) AddEntry(uri string) {
	rl.trim()

	newEntry := LogEntry{url: uri,
		timestamp: time.Now()}
	rl.entries.PushBack(newEntry)
}

// trims the entries at the start of the list if they fall before the cutoff duration
func (rl *RollingLog) trim() {
	lastValidTime := time.Now().Add(-rl.keepfor)

	e := rl.entries.Front()
	for e != nil && e.Value.(LogEntry).timestamp.Before(lastValidTime) {
		next := e.Next()
		rl.entries.Remove(e)
		e = next
	}
}

// GetAllCounts - returns a map with all the count of values in the rolling log
func (rl *RollingLog) GetAllCounts() map[string]int {
	rl.trim()

	result := make(map[string]int)

	for e := rl.entries.Front(); e != nil; e = e.Next() {
		key := e.Value.(LogEntry).url
		elem, ok := result[key]
		if ok {
			elem = elem + 1
		} else {
			elem = 1
		}
		result[key] = elem
	}

	return result

}
