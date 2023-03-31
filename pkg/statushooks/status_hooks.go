package statushooks

type StatusHooks interface {
	SetStatus(string)
	Show()
	Warn(string)
	Hide()
	Message(...string)
}
