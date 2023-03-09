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

// JsonCloneToMap tries to JSON marshal and unmarshal the given data, returning a map[string]any if successful
func JsonCloneToMap(val any) (map[string]any, error) {
	var res map[string]any
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonBytes, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
