package controldisplay

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"time"
)

// toCsvCell escapes a value for csv
func toCsvCell(v reflect.Value) string {
	buffer := bytes.NewBufferString("")
	csvWriter := csv.NewWriter(buffer)
	csvWriter.Write([]string{fmt.Sprintf("\"%v\"", v)})
	return buffer.String()
}

// durationInSeconds returns the passed in duration as seconds
func durationInSeconds(t time.Duration) float64 { return t.Seconds() }
