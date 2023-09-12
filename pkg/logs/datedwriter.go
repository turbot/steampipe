package logs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

type DatedWriter struct {
	directory string
	prefix    string

	currentWriter io.Writer
	currentPath   string
}

func NewDatedWriter(directory string, prefix string) *DatedWriter {
	return &DatedWriter{
		directory: directory,
		prefix:    prefix,
	}
}

func (dwr *DatedWriter) Write(p []byte) (n int, err error) {
	pathShouldBe := filepath.Join(dwr.directory, fmt.Sprintf("%s-%s.log", dwr.prefix, time.Now().Format(time.DateOnly)))
	if dwr.currentPath != pathShouldBe {
		// we need to flush the current one
		// try to cast it to a Closer (if this is nil, isCloseable will be false)
		closeableWriter, isCloseable := dwr.currentWriter.(io.Closer)
		if isCloseable {
			closeableWriter.Close()
		}
		// create a new one
		dwr.currentPath = pathShouldBe
		f, err := os.OpenFile(dwr.currentPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return 0, sperr.WrapWithRootMessage(err, "failed to open steampipe log file")
		}
		dwr.currentWriter = f
	}

	return dwr.currentWriter.Write(p)
}
