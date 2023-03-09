package display

import "fmt"

func ClearCurrentLine() {
	fmt.Print("\n\033[1A\033[K")
}
