package initialisation

type InitStatusHook struct {
	initData *InitData
}

func NewInitStatusHook(initData *InitData) *InitStatusHook {
	hooks := new(InitStatusHook)
	hooks.initData = initData
	return hooks
}

func (h *InitStatusHook) SetStatus(arg string) {
	h.initData.SetStatus(arg)
}
func (h *InitStatusHook) Done()             {}
func (h *InitStatusHook) Message(...string) {}
