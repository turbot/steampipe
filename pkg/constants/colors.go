package constants

import (
	"github.com/logrusorgru/aurora"
)

// Colors is a map of string to aurora colour value
var Colors = map[string]func(arg interface{}) aurora.Value{
	"bold":       Bold,
	"italic":     Italic,
	"underline":  Underline,
	"slow-blink": SlowBlink,

	"black":   Black,
	"red":     Red,
	"green":   Green,
	"yellow":  Yellow,
	"blue":    Blue,
	"magenta": Magenta,
	"cyan":    Cyan,
	"white":   White,

	"bold-black":   BoldBlack,
	"bold-red":     BoldRed,
	"bold-green":   BoldGreen,
	"bold-yellow":  BoldYellow,
	"bold-blue":    BoldBlue,
	"bold-magenta": BoldMagenta,
	"bold-cyan":    BoldCyan,
	"bold-white":   BoldWhite,

	"bright-black":   BrightBlack,
	"bright-red":     BrightRed,
	"bright-green":   BrightGreen,
	"bright-yellow":  BrightYellow,
	"bright-blue":    BrightBlue,
	"bright-magenta": BrightMagenta,
	"bright-cyan":    BrightCyan,
	"bright-white":   BrightWhite,

	"bold-bright-black":   BoldBrightBlack,
	"bold-bright-red":     BoldBrightRed,
	"bold-bright-green":   BoldBrightGreen,
	"bold-bright-yellow":  BoldBrightYellow,
	"bold-bright-blue":    BoldBrightBlue,
	"bold-bright-magenta": BoldBrightMagenta,
	"bold-bright-cyan":    BoldBrightCyan,
	"bold-bright-white":   BoldBrightWhite,

	"gray1": Gray1,
	"gray2": Gray2,
	"gray3": Gray3,
	"gray4": Gray4,
	"gray5": Gray5,
}

var Bold = aurora.Bold
var Italic = aurora.Italic
var Underline = aurora.Underline
var SlowBlink = aurora.SlowBlink
var Blink = aurora.Blink

var Black = aurora.Black
var Red = aurora.Red
var Green = aurora.Green
var Yellow = aurora.Yellow
var Blue = aurora.Blue
var Magenta = aurora.Magenta
var Cyan = aurora.Cyan
var White = aurora.White

// bright colors

var BrightBlack = aurora.BrightBlack
var BrightRed = aurora.BrightRed
var BrightGreen = aurora.BrightGreen
var BrightYellow = aurora.BrightYellow
var BrightBlue = aurora.BrightBlue
var BrightMagenta = aurora.BrightMagenta
var BrightCyan = aurora.BrightCyan
var BrightWhite = aurora.BrightWhite

// bold colors

func BoldBlack(arg interface{}) aurora.Value {
	return Bold(Black(arg))
}
func BoldRed(arg interface{}) aurora.Value {
	return Bold(Red(arg))
}
func BoldGreen(arg interface{}) aurora.Value {
	return Bold(Green(arg))
}
func BoldYellow(arg interface{}) aurora.Value {
	return Bold(Yellow(arg))
}
func BoldBlue(arg interface{}) aurora.Value {
	return Bold(Blue(arg))
}
func BoldMagenta(arg interface{}) aurora.Value {
	return Bold(Magenta(arg))
}
func BoldCyan(arg interface{}) aurora.Value {
	return Bold(Cyan(arg))
}
func BoldWhite(arg interface{}) aurora.Value {
	return Bold(White(arg))
}

// bold bright  colors

func BoldBrightBlack(arg interface{}) aurora.Value {
	return Bold(BrightBlack(arg))
}
func BoldBrightRed(arg interface{}) aurora.Value {
	return Bold(BrightRed(arg))
}
func BoldBrightGreen(arg interface{}) aurora.Value {
	return Bold(BrightGreen(arg))
}
func BoldBrightYellow(arg interface{}) aurora.Value {
	return Bold(BrightYellow(arg))
}
func BoldBrightBlue(arg interface{}) aurora.Value {
	return Bold(BrightBlue(arg))
}
func BoldBrightMagenta(arg interface{}) aurora.Value {
	return Bold(BrightMagenta(arg))
}
func BoldBrightCyan(arg interface{}) aurora.Value {
	return Bold(BrightCyan(arg))
}
func BoldBrightWhite(arg interface{}) aurora.Value {
	return Bold(BrightWhite(arg))
}

// various preset grays - lower number is a darker grey

func Gray1(arg interface{}) aurora.Value {
	return aurora.Gray(6, arg)
}

func Gray2(arg interface{}) aurora.Value {
	return aurora.Gray(10, arg)
}

func Gray3(arg interface{}) aurora.Value {
	return aurora.Gray(14, arg)
}

func Gray4(arg interface{}) aurora.Value {
	return aurora.Gray(16, arg)
}

func Gray5(arg interface{}) aurora.Value {
	return aurora.Gray(20, arg)
}
