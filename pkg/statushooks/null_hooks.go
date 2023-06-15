package statushooks

var NullHooks StatusHooks = &NullStatusHook{}

type NullStatusHook struct{}

func (*NullStatusHook) SetStatus(string)  {}
func (*NullStatusHook) Hide()             {}
func (*NullStatusHook) Message(...string) {}
func (*NullStatusHook) Show()             {}
func (*NullStatusHook) Warn(string)       {}
