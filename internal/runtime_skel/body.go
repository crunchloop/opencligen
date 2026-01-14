package runtime

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// LoadBody loads request body from a data string
// Supports:
// - @filename - reads from file
// - @- - reads from stdin
// - raw JSON string
func LoadBody(data string) ([]byte, error) {
	if data == "" {
		return nil, nil
	}

	// Check for file reference
	if strings.HasPrefix(data, "@") {
		path := data[1:]

		if path == "-" {
			// Read from stdin
			return io.ReadAll(os.Stdin)
		}

		// Read from file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read body from file %s: %w", path, err)
		}
		return content, nil
	}

	// Treat as raw JSON
	return []byte(data), nil
}
