package controldisplay

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
)

// templateFuncs merges desired functions from sprig with custom functions that we
// define in steampipe
func templateFuncs() template.FuncMap {
	useFromSprigMap := []string{"upper", "toJson", "quote", "dict", "add", "now", "toPrettyJson"}

	var funcs template.FuncMap = template.FuncMap{}
	sprigMap := sprig.TxtFuncMap()
	for _, use := range useFromSprigMap {
		f, found := sprigMap[use]
		if !found {
			// guarantee that when a function is expected to be present
			// it does not slip through any crack
			panic(fmt.Sprintf("%s not found", use))
		}
		if found {
			funcs[use] = f
		}
	}
	for k, v := range formatterTemplateFuncMap {
		funcs[k] = v
	}

	return funcs
}

// custom steampipe functions - ones we couldn't find in sprig
var formatterTemplateFuncMap template.FuncMap = template.FuncMap{
	"durationInSeconds": durationInSeconds,
	"toCsvCell":         toCsvCell,
}

var (
	csvWriterBuffer = bytes.NewBufferString("")
	csvWriter       = csv.NewWriter(csvWriterBuffer)
	csvBufferLock   = sync.Mutex{}
)

// toCsvCell escapes a value for csv
func toCsvCell(v interface{}) string {
	csvBufferLock.Lock()
	defer csvBufferLock.Unlock()

	csvWriterBuffer.Reset()
	csvWriter.Write([]string{fmt.Sprintf("%v", v)})
	csvWriter.Flush()
	return strings.TrimSpace(csvWriterBuffer.String())
}

// durationInSeconds returns the passed in duration as seconds
func durationInSeconds(t time.Duration) float64 { return t.Seconds() }
