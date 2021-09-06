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

// Push :: add to the history queue; trim to maxHistorySize if necessary
func (q *QueryHistory) Push(query string) {

	// do a strict compare to see if we have this same exact query as the most recent history item
	if len(q.history) > 0 && q.history[len(q.history)-1] == query {
		return
	}

	// limit the history length to HistorySize
	historyLength := len(q.history)
	if historyLength >= constants.HistorySize {
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
