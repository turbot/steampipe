package statushooks

type StatusHooks interface {
	SetStatus(string)
	Done()
	Message(...string)
}
