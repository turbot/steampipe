package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/turbot/steampipe/pkg/version"
)

func getUserAgent() string {
	return fmt.Sprintf("Turbot Steampipe/%s (+https://steampipe.io)", version.SteampipeVersion.String())
}

// BuildRequestPayload merges the provided payload with the standard payload that needs to be sent
func BuildRequestPayload(signature string, payload map[string]interface{}) *bytes.Buffer {
	requestPayload := map[string]interface{}{
		"version":     version.SteampipeVersion.String(),
		"os_platform": runtime.GOOS,
		"arch":        runtime.GOARCH,
		"signature":   signature,
	}

	// change the platform to "windows_linux" if we are running in "Windows Subsystem for Linux"
	if runtime.GOOS == "linux" {
		if IsWSL() {
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

// SendRequest makes a http call to the given URL
func SendRequest(ctx context.Context, signature string, method string, sendRequestTo url.URL, payload io.Reader) (*http.Response, error) {
	// Set a default timeout of 3 sec for the check request (in milliseconds)
	req, err := http.NewRequestWithContext(ctx, method, sendRequestTo.String(), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", getUserAgent())

	client := cleanhttp.DefaultClient()

	return client.Do(req)
}
