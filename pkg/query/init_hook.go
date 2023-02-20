package query

type QueryInitStatusHook struct {
	initData *InitData
}

func NewQueryInitStatusHook(initData *InitData) *QueryInitStatusHook {
	hooks := new(QueryInitStatusHook)
	hooks.initData = initData
	return hooks
}

func (h *QueryInitStatusHook) SetStatus(status string) {
	h.initData.SetStatus(status)
}
func (h *QueryInitStatusHook) Done()             {}
func (h *QueryInitStatusHook) Message(...string) {}
