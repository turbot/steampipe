package queryresult

type ResultStreamer struct {
	Results      chan *Result
	displayReady chan string
	started      bool
}

func NewResultStreamer() *ResultStreamer {
	return &ResultStreamer{
		// make buffered channel  so we can always stream a single result
		Results:      make(chan *Result, 1),
		displayReady: make(chan string, 1),
	}
}

func (q *ResultStreamer) StreamResult(result *Result) {
	q.Results <- result
}

func (q *ResultStreamer) StreamSingleResult(result *Result) {
	q.Results <- result
	q.Wait()
	close(q.Results)
}

func (q *ResultStreamer) Start() {
	q.started = true
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
	if q.started {
		<-q.displayReady
		q.started = false
	}
}
