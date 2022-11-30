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
func templateFuncs(renderContext TemplateRenderContext) template.FuncMap {
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
	// custom steampipe functions - ones we couldn't find in sprig
	formatterTemplateFuncMap := template.FuncMap{
		"durationInSeconds": durationInSeconds,
		"toCsvCell":         toCSVCellFnFactory(renderContext.Config.Separator),
	}
	for k, v := range formatterTemplateFuncMap {
		funcs[k] = v
	}

	return funcs
}

// toCsvCell escapes a value for csv
// we need to do this in a factory function, so that we can
// set the separator for the CSV writer for this render session
func toCSVCellFnFactory(comma string) func(interface{}) string {
	csvWriterBuffer := bytes.NewBufferString("")
	csvBufferLock := sync.Mutex{}

	csvWriter := csv.NewWriter(csvWriterBuffer)
	if len(comma) > 0 {
		csvWriter.Comma = []rune(comma)[0]
	}

	return func(v interface{}) string {
		csvBufferLock.Lock()
		defer csvBufferLock.Unlock()

		csvWriterBuffer.Reset()
		csvWriter.Write([]string{fmt.Sprintf("%v", v)})
		csvWriter.Flush()
		return strings.TrimSpace(csvWriterBuffer.String())
	}
}

// durationInSeconds returns the passed in duration as seconds
func durationInSeconds(t time.Duration) float64 { return t.Seconds() }
