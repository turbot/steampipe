package queryresult

type ResultStreamer struct {
	Error              chan error
	Results            chan *Result
	allResultsReceived chan string
}

func NewResultStreamer() *ResultStreamer {
	return &ResultStreamer{
		// make buffered channel so we can always stream a single result
		Error:              make(chan error, 1),
		Results:            make(chan *Result, 1),
		allResultsReceived: make(chan string, 1),
	}
}

// StreamResult streams result on the Results channel, then waits for them to be read
func (q *ResultStreamer) StreamResult(result *Result) {
	q.Results <- result
	// wait for the result to be read
	<-q.allResultsReceived
}

// StreamError sets an error
func (q *ResultStreamer) StreamError(err error) {
	q.Error <- err
}

// Close closes the result stream
func (q *ResultStreamer) Close() {
	close(q.Results)
}

// AllResultsRead is a signal that indicates the all results have been read from the stream
func (q *ResultStreamer) AllResultsRead() {
	q.allResultsReceived <- ""
}
