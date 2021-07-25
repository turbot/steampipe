package db_common

import "fmt"

type InitResult struct {
	Error    error
	Warnings []string
	Messages []string
}

func (r *InitResult) AddMessage(message string) {
	r.Messages = append(r.Messages, message)
}

func (r *InitResult) AddWarnings(warnings []string) {
	r.Warnings = append(r.Warnings, warnings...)
}

func (r *InitResult) HasMessages() bool {
	return len(r.Warnings)+len(r.Messages) > 0
}

func (r *InitResult) DisplayMessages() {
	for _, w := range r.Warnings {
		fmt.Println(w)
	}
	for _, w := range r.Messages {
		fmt.Println(w)
	}
}
