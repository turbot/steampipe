package db

type ResultStreamer struct {
	Results      chan *QueryResult
	displayReady chan string
}

func newQueryResults() *ResultStreamer {
	return &ResultStreamer{
		// make buffered channel  so we can always stream a single result
		Results:      make(chan *QueryResult, 1),
		displayReady: make(chan string, 1),
	}
}

func (q *ResultStreamer) streamResult(result *QueryResult) {
	q.Results <- result
}

func (q *ResultStreamer) streamSingleResult(result *QueryResult, onComplete func()) {
	q.Results <- result
	q.Wait()
	onComplete()
	close(q.Results)
}

func (q *ResultStreamer) close() {
	close(q.Results)
}

// Done :: signals that the next QueryResult has been processed
func (q *ResultStreamer) Done() {
	q.displayReady <- ""
}

// Wait :: waits for the next QueryResult to get processed
func (q *ResultStreamer) Wait() {
	<-q.displayReady
}
