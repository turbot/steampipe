package statushooks

import "fmt"

var ConsoleHook = &ConsoleLogStatusHook{}

type ConsoleLogStatusHook struct{}

func (*ConsoleLogStatusHook) SetStatus(msg string) {
	fmt.Printf("[ STATUS  ] %s\n", msg)
}
func (*ConsoleLogStatusHook) Message(msgs ...string) {
	for _, msg := range msgs {
		fmt.Printf("[ MESSAGE ] %s\n", msg)
	}
}
func (*ConsoleLogStatusHook) Done() {}
