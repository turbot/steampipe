package constants

import (
	"github.com/fatih/color"
)

var Bold = color.New(color.Bold).SprintFunc()
var Red = color.New(color.FgRed).SprintFunc()
var BoldYellow = color.New(color.Bold, color.FgYellow).SprintFunc()
