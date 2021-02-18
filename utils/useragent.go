package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/containerd/containerd/version"
	"github.com/hashicorp/go-cleanhttp"
)

func ConstructUserAgent(installationID string) string {
	return fmt.Sprintf("Turbot Steampipe/%s (+https://steampipe.io)", version.Version)
}

// BuildRequestPayload :: merges the provided payload with the standard payload that needs to be sent
func BuildRequestPayload(signature string, payload map[string]interface{}) *bytes.Buffer {
	requestPayload := map[string]interface{}{
		"version":     version.Version,
		"os_platform": runtime.GOOS,
		"arch":        runtime.GOARCH,
		"signature":   signature,
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
	timeout := 3000 * time.Millisecond
	req, err := http.NewRequest(method, sendRequestTo.String(), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", ConstructUserAgent(signature))

	client := cleanhttp.DefaultClient()

	// Use a short timeout since checking for new versions is not critical
	// enough to block on if the update server is broken/slow.
	client.Timeout = timeout

	return client.Do(req)
}
