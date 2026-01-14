package runtime

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// MaxSSEEventSize is the maximum allowed size for a single SSE event (10MB)
const MaxSSEEventSize = 10 * 1024 * 1024

// ErrSSEEventTooLarge is returned when an SSE event exceeds MaxSSEEventSize
var ErrSSEEventTooLarge = errors.New("SSE event data exceeds maximum allowed size")

// isEventStream checks if content type indicates SSE
func isEventStream(contentType string) bool {
	return strings.Contains(contentType, "text/event-stream")
}

// handleSSE handles Server-Sent Events response
func handleSSE(reader io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(reader)
	var dataBuffer strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" {
			// Empty line signals end of event
			if dataBuffer.Len() > 0 {
				data := strings.TrimSpace(dataBuffer.String())
				if data != "" {
					// Print the data (typically JSON)
					if isJSON([]byte(data)) {
						prettyPrint([]byte(data), out)
					} else {
						fmt.Fprintln(out, data)
					}
				}
				dataBuffer.Reset()
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			// Comment/keep-alive, skip
			continue
		}

		if strings.HasPrefix(line, "data:") {
			// Extract data after "data:"
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)

			// Check buffer size limit before appending
			newSize := dataBuffer.Len() + len(data) + 1 // +1 for potential newline
			if newSize > MaxSSEEventSize {
				return ErrSSEEventTooLarge
			}

			if dataBuffer.Len() > 0 {
				dataBuffer.WriteString("\n")
			}
			dataBuffer.WriteString(data)
			continue
		}

		// Handle other SSE fields (event, id, retry) - we just skip them for now
		if strings.HasPrefix(line, "event:") ||
			strings.HasPrefix(line, "id:") ||
			strings.HasPrefix(line, "retry:") {
			continue
		}
	}

	// Handle any remaining data
	if dataBuffer.Len() > 0 {
		data := strings.TrimSpace(dataBuffer.String())
		if data != "" {
			if isJSON([]byte(data)) {
				prettyPrint([]byte(data), out)
			} else {
				fmt.Fprintln(out, data)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading SSE stream: %w", err)
	}

	return nil
}
