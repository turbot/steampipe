package queryresult

type ResultStreamer struct {
	Results      chan *Result
	displayReady chan string
	streaming    bool
}

func NewResultStreamer() *ResultStreamer {
	return &ResultStreamer{
		// make buffered channel  so we can always stream a single result
		Results:      make(chan *Result, 1),
		displayReady: make(chan string, 1),
	}
}

func (q *ResultStreamer) StreamResult(result *Result) {
	q.streaming = true
	q.Results <- result
}

func (q *ResultStreamer) StreamSingleResult(result *Result) {
	q.streaming = true
	q.Results <- result
	q.Wait()
	close(q.Results)
}

func (q *ResultStreamer) Close() {
	close(q.Results)
}

// Done :: signals that the next Result has been processed
func (q *ResultStreamer) Done() {
	q.displayReady <- ""
}

// Wait :: waits for the next Result to get processed
func (q *ResultStreamer) Wait() {
	if q.streaming {
		<-q.displayReady
	}
	q.streaming = false
}
