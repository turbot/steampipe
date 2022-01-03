package statushooks

type StatusHooks interface {
	SetStatus(string)
	Done()
	//SetStatusAfterDelay(string, time.Duration, chan bool)
}

var Null = &NullStatusHook{}

type NullStatusHook struct{}

func (*NullStatusHook) SetStatus(string) {}
func (*NullStatusHook) Done()            {}
