package statushooks

var NullHooks = &NullStatusHook{}

type NullStatusHook struct{}

func (*NullStatusHook) SetStatus(string)  {}
func (*NullStatusHook) Done()             {}
func (*NullStatusHook) Message(...string) {}
