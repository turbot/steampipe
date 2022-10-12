package utils

import (
	"bytes"
	"encoding/json"
	"io"
)

func PrettifyJsonFromReader(dataToExport io.Reader) (io.Reader, error) {
	b, err := io.ReadAll(dataToExport)
	if err != nil {
		return nil, err
	}
	var prettyJSON bytes.Buffer

	err = json.Indent(&prettyJSON, b, "", "  ")
	if err != nil {
		return nil, err
	}
	dataToExport = &prettyJSON
	return dataToExport, nil
}
