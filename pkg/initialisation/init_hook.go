package initialisation

type InitStatusHook struct {
	initData *InitData
}

func NewInitStatusHook(initData *InitData) *InitStatusHook {
	hooks := new(InitStatusHook)
	hooks.initData = initData
	return hooks
}

func (h *InitStatusHook) SetStatus(status string) {
	h.initData.SetStatus(status)
}
func (h *InitStatusHook) Done()             {}
func (h *InitStatusHook) Message(...string) {}
