package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// handleResponse handles a standard HTTP response
func handleResponse(resp *http.Response, out io.Writer) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-2xx status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "Error: HTTP %d %s\n", resp.StatusCode, resp.Status)
		if len(body) > 0 {
			fmt.Fprintln(os.Stderr, string(body))
		}
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// Pretty print JSON if possible
	if len(body) > 0 {
		if isJSON(body) {
			prettyPrint(body, out)
		} else {
			fmt.Fprintln(out, string(body))
		}
	}

	return nil
}

// isJSON checks if the content is valid JSON
func isJSON(data []byte) bool {
	var js interface{}
	return json.Unmarshal(data, &js) == nil
}

// prettyPrint outputs JSON with indentation
func prettyPrint(data []byte, out io.Writer) {
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		// Fall back to raw output
		fmt.Fprintln(out, string(data))
		return
	}

	pretty, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		fmt.Fprintln(out, string(data))
		return
	}

	fmt.Fprintln(out, string(pretty))
}
