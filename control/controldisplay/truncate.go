package controldisplay

import "fmt"

func TruncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return fmt.Sprintf("%s…", str[:length-1])
}
