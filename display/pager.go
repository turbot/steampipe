package display

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/karrick/gows"
)

func displayPaged(content string) {
	if isPagerNeeded(content) {
		switch runtime.GOOS {
		case "darwin", "linux":
			nixLessPager(content)
		case "windows":
			winMorePager(content)
		default:
			nullPager(content)
		}
	} else {
		nullPager(content)
	}
}

func isPagerNeeded(content string) bool {
	maxCols, maxRow, _ := gows.GetWinSize()

	// let's scan through it instead of iterating over it fully
	sc := bufio.NewScanner(strings.NewReader(content))
	lineCount := 0
	for sc.Scan() {
		line := sc.Text()
		lineCount++
		if lineCount > maxRow {
			return true
		}
		if len(line) > maxCols {
			return true
		}
	}
	return false
}

func winMorePager(content string) {
	// for the time being, route this to the nullpager
	// eventually use windows `more` with a temp file
	nullPager(content)
}

func nullPager(content string) {
	// just dump the whole thing out
	// we will use this for non-tty environments as well
	fmt.Print(content)
}

func nixLessPager(content string) {
	lessProcess := exec.Command("less", "-SRXF")
	lessProcess.Stdout = os.Stdout
	lessProcess.Stderr = os.Stderr
	lessProcess.Stdin = strings.NewReader(content)
	// Run it, so that this blocks out the go-prompt stuff.
	// No point Start-ing it anyway
	lessProcess.Run()
}
