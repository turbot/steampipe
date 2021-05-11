package controlresult

// StatusSummary is a struct containing the counts of each possible control status
type StatusSummary struct {
	Alarm int
	Ok    int
	Info  int
	Skip  int
	Error int
}
