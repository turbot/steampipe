package queryhistory

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/constants"
)

// QueryHistory :: struct for working with history in the interactive mode
type QueryHistory struct {
	history []string
}

// New :: create and return a
func New() *QueryHistory {
	history := new(QueryHistory)
	history.load()
	return history
}

// Put :: add to the history queue; trim to maxHistorySize if necessary
func (q *QueryHistory) Put(query string) {
	// if !cmdconfig.Viper().GetBool(constants.ArgMultiLine) || metaquery.IsMetaQuery(query) {
	// 	q.history = append(q.history, query)
	// 	return
	// }

	// this is a multi line query

	// do a strict compare to see if we have this same exact one
	// somewhere in history
	idx, found := q.find(query)
	if found {
		// helpers.RemoveFromStringSlice() is going to be slow, since it
		// iterates over the array, this is probably going to
		// be faster, especially since we know the index we want to remove
		q.removeEntryAt(idx)
	}

	// check the current length
	historyLength := len(q.history)
	if historyLength >= constants.HistorySize {
		// trim out the last HistorySize elements
		q.history = q.history[historyLength-constants.HistorySize+1:]
	}

	// append the new entry
	q.history = append(q.history, query)
}

// Persist :: persist the history to the filesystem
func (q *QueryHistory) Persist() error {
	var file *os.File
	var err error
	defer func() {
		file.Close()
	}()
	path := filepath.Join(constants.InternalDir(), constants.HistoryFile)
	file, err = os.Create(path)
	if err != nil {
		return err
	}

	jsonEncoder := json.NewEncoder(file)

	// disable indentation
	jsonEncoder.SetIndent("", "")

	return jsonEncoder.Encode(q.history)
}

// Get :: return a copy of the current history
func (q *QueryHistory) Get() []string {
	return q.history
}

// loads up the history from the file where it is persisted
func (q *QueryHistory) load() error {
	path := filepath.Join(constants.InternalDir(), constants.HistoryFile)
	file, err := os.Open(path)
	if err != nil {
		q.history = []string{}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&q.history)
}

func (q *QueryHistory) find(val string) (int, bool) {
	for i, item := range q.history {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func (q *QueryHistory) removeEntryAt(idx int) {
	if idx == 0 {
		q.history = q.history[1:]
	}
	q.history = append(
		q.history[:idx],
		q.history[idx+1:]...,
	)
}
