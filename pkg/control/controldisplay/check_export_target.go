package controldisplay

type CheckExportTarget struct {
	Formatter Formatter
	File      string
}

func NewCheckExportTarget(formatter Formatter, file string) CheckExportTarget {
	return CheckExportTarget{
		Formatter: formatter,
		File:      file,
	}
}
