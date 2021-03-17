package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// functions specifically used for Debugging purposes.
func DebugDumpJSON(msg string, d interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent(" ", " ")
	os.Stdout.WriteString(msg)
	enc.Encode(d)
}

func DebugDumpViper() {
	fmt.Println(strings.Repeat("*", 80))
	for _, vKey := range viper.AllKeys() {
		fmt.Printf("%-30s => %v\n", vKey, viper.Get(vKey))
	}
	fmt.Println(strings.Repeat("*", 80))
}
