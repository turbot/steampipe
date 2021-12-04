package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/turbot/steampipe/version"
)

const httpTimeout = 5 * time.Second

func getUserAgent() string {
	return fmt.Sprintf("Turbot Steampipe/%s (+https://steampipe.io)", version.SteampipeVersion.String())
}

// BuildRequestPayload :: merges the provided payload with the standard payload that needs to be sent
func BuildRequestPayload(signature string, payload map[string]interface{}) *bytes.Buffer {
	requestPayload := map[string]interface{}{
		"version":     version.SteampipeVersion.String(),
		"os_platform": runtime.GOOS,
		"arch":        runtime.GOARCH,
		"signature":   signature,
	}

	// change the platform to "windows_linux" if we are running in "Windows Subsystem for Linux"
	if runtime.GOOS == "linux" {
		wsl, err := IsWSL()
		if err == nil && wsl {
			requestPayload["os_platform"] = "windows_linux"
		}
	}

	// now merge the given payload
	for k, v := range payload {
		_, alreadyThere := requestPayload[k]
		if alreadyThere {
			panic("cannot merge already existing properties")
		}
		requestPayload[k] = v
	}

	jsonStr, _ := json.Marshal(requestPayload)
	return bytes.NewBuffer(jsonStr)
}

// SendRequest ::
func SendRequest(signature string, method string, sendRequestTo url.URL, payload *bytes.Buffer) (*http.Response, error) {
	// Set a default timeout of 3 sec for the check request (in milliseconds)
	req, err := http.NewRequest(method, sendRequestTo.String(), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", getUserAgent())

	client := cleanhttp.DefaultClient()

	// Use a short timeout since checking for new versions is not critical
	// enough to block on if the update server is broken/slow.
	client.Timeout = httpTimeout

	log.Println("[TRACE]", "Sending HTTP Request", req)

	return client.Do(req)
}
