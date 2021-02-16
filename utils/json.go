package utils

import (
	"encoding/json"
	"os"
)

func DebugSpitJSON(msg string, d interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent(" ", " ")
	os.Stdout.WriteString(msg)
	enc.Encode(d)
}
