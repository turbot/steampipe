package utils

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
)

var (
	colorErr    = color.RedString("Error")
	colorWarn   = color.YellowString("Warning")
	colorNotice = color.GreenString("Notice")
)

func init() {
	color.Output = ansicolor.NewAnsiColorWriter(os.Stderr)
}

func FailOnError(err error) {
	if err != nil {
		panic(err)
	}
}
func FailOnErrorWithMessage(err error, message string) {
	if err != nil {
		panic(fmt.Sprintf("%s %s", message, err.Error()))
	}
}

func ShowError(err error) {
        errString := strings.TrimSpace(err.Error())
	if strings.HasPrefix(errString, "pq:") {
		errString = strings.TrimSpace(strings.TrimPrefix(errString, "pq:"))
		if strings.HasPrefix(errString, "rpc error") {
			errString = errString[33:]
		}
	}
	fmt.Fprintf(color.Output, "%s: %v\n", colorErr, errString)
}

func ShowErrorWithMessage(err error, message string) {
	errString := strings.TrimSpace(err.Error())
	if strings.HasPrefix(errString, "pq:") {
		errString = strings.TrimSpace(strings.TrimPrefix(errString, "pq:"))
		if strings.HasPrefix(errString, "rpc error") {
			errString = errString[33:]
		}
	}
        fmt.Fprintf(color.Output, "%s: %s - %v\n", colorErr, message, errString)
}

func ShowWarning(warning string) {
	fmt.Fprintf(color.Output, "%s: %v\n", colorWarn, warning)
}
