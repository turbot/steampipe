package queryhistory

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
)

// QueryHistory :: struct for working with history in the interactive mode
type QueryHistory struct {
	history []string
}

// New creates a new QueryHistory object
func New() (*QueryHistory, error) {
	history := &QueryHistory{history: []string{}}
	err := history.load()
	if err != nil {
		return nil, err
	}
	return history, nil
}

// Push adds a string to the history queue trimming to maxHistorySize if necessary
func (q *QueryHistory) Push(query string) {
	if len(strings.TrimSpace(query)) == 0 {
		// do not store a blank query
		return
	}

	// do a strict compare to see if we have this same exact query as the most recent history item
	if lastElement := q.Peek(); lastElement != nil && (*lastElement) == query {
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

// Peek returns the last element of the history stack.
// returns nil if there is no history
func (q *QueryHistory) Peek() *string {
	if len(q.history) == 0 {
		return nil
	}
	return &q.history[len(q.history)-1]
}

// Persist writes the history to the filesystem
func (q *QueryHistory) Persist() error {
	var file *os.File
	var err error
	defer func() {
		file.Close()
	}()
	path := filepath.Join(filepaths.EnsureInternalDir(), constants.HistoryFile)
	file, err = os.Create(path)
	if err != nil {
		return err
	}

	jsonEncoder := json.NewEncoder(file)

	// disable indentation
	jsonEncoder.SetIndent("", "")

	return jsonEncoder.Encode(q.history)
}

// Get returns the full history
func (q *QueryHistory) Get() []string {
	return q.history
}

// loads up the history from the file where it is persisted
func (q *QueryHistory) load() error {
	path := filepath.Join(filepaths.EnsureInternalDir(), constants.HistoryFile)
	file, err := os.Open(path)
	if err != nil {
		// ignore not exists errors
		if os.IsNotExist(err) {
			return nil
		}
		return err

	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&q.history)
	// ignore EOF (caused by empty file)
	if err == io.EOF {
		return nil
	}
	return err
}
