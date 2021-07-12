package db

import "fmt"

type InitResult struct {
	Error    error
	Warnings []string
	Messages []string
}

func (d *InitResult) AddMessage(message string) {
	d.Messages = append(d.Messages, message)
}

func (d *InitResult) AddWarning(warning string) {
	d.Warnings = append(d.Warnings, warning)
}

func (d *InitResult) DisplayWarnings() {
	for _, w := range d.Warnings {
		fmt.Println(w)
	}
}
func (d *InitResult) DisplayMessages() {
	for _, w := range d.Messages {
		fmt.Println(w)
	}
}
