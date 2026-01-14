package plan

import (
	"regexp"
	"strings"
	"unicode"
)

// DeriveCommandName derives a command name from an operationId
func DeriveCommandName(operationID string) string {
	// Common prefix patterns to simplify
	prefixes := []struct {
		pattern string
		replace string
	}{
		{`^list[A-Z]`, "list"},
		{`^get[A-Z]`, "get"},
		{`^create[A-Z]`, "create"},
		{`^update[A-Z]`, "update"},
		{`^delete[A-Z]`, "delete"},
		{`^start[A-Z]`, "start"},
		{`^stop[A-Z]`, "stop"},
		{`^cancel[A-Z]`, "cancel"},
		{`^ping[A-Z]`, "ping"},
		{`^subscribe[A-Z]`, "subscribe"},
	}

	for _, p := range prefixes {
		re := regexp.MustCompile(p.pattern)
		if re.MatchString(operationID) {
			return p.replace
		}
	}

	// Fallback: convert to kebab-case
	return toKebabCase(operationID)
}

// DeriveGroupName derives a group name from a tag
func DeriveGroupName(tag string) string {
	if tag == "" {
		return "default"
	}
	return toKebabCase(tag)
}

// DeriveFlagName derives a flag name from a parameter name
func DeriveFlagName(paramName, in string) string {
	name := paramName

	// Strip X- prefix for headers
	if in == "header" && strings.HasPrefix(strings.ToUpper(name), "X-") {
		name = name[2:]
	}

	return toKebabCase(name)
}

// toKebabCase converts a string to kebab-case
func toKebabCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	for i, r := range s {
		switch {
		case unicode.IsUpper(r):
			if i > 0 {
				// Don't add hyphen if previous char was already uppercase (e.g., "ID" -> "id")
				prev := rune(s[i-1])
				if !unicode.IsUpper(prev) && prev != '-' && prev != '_' {
					result.WriteRune('-')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		case r == '_' || r == ' ':
			result.WriteRune('-')
		default:
			result.WriteRune(r)
		}
	}

	// Clean up any double hyphens
	res := result.String()
	for strings.Contains(res, "--") {
		res = strings.ReplaceAll(res, "--", "-")
	}

	return strings.Trim(res, "-")
}

// ParseCommandPath parses a space-delimited command path
// e.g., "tasks activities" -> ["tasks", "activities"]
func ParseCommandPath(name string) []string {
	parts := strings.Fields(name)
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = toKebabCase(p)
	}
	return result
}
