package db

type InitResult struct {
	Error    error
	Warnings []string
	Messages []string
}

func (d *InitResult) Merge(other *InitResult) {
	// if there is already an error, od not overwrite - retain the first one to occur
	if d.Error == nil && other.Error != nil {
		d.Error = other.Error
	}
	d.Warnings = append(d.Warnings, other.Warnings...)
	d.Messages = append(d.Messages, other.Messages...)
}

func (d *InitResult) AddMessage(message string) {
	d.Messages = append(d.Messages, message)
}

func (d *InitResult) AddWarning(warning string) {
	d.Messages = append(d.Warnings, warning)
}
