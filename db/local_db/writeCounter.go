package local_db

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/karrick/gows"
)

// struct for a sort of a progress indicator when downloading the JAR
type writeCounter struct {
	title string
	total uint64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc *writeCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	cols, _, _ := gows.GetWinSize()
	// to clear out the line
	fmt.Printf("\r%s", strings.Repeat(" ", cols))
	// print the update
	fmt.Printf("\rDownloading %s ... ~%s complete", wc.title, humanize.Bytes(wc.total))
}
