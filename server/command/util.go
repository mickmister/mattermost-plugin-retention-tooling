package command

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	SubCommandKey = "-subcommand"
)

// parseNamedArgs parses a command string into a map of arguments. It is assumed the
// command string is of the form `<subcommand> --arg1 value1 ...` Supports empty values.
// Arg names are limited to [0-9a-zA-Z_].
func parseNamedArgs(cmd string) map[string]string {
	m := make(map[string]string)

	split := strings.Fields(cmd)

	// check for optional action
	if len(split) >= 2 && !strings.HasPrefix(split[1], "--") {
		m[SubCommandKey] = split[1] // prefix with hyphen to avoid collision with arg named "subcommand"
	}

	for i := 0; i < len(split); i++ {
		if !strings.HasPrefix(split[i], "--") {
			continue
		}
		var val string
		arg := trimSpaceAndQuotes(strings.Trim(split[i], "-"))
		if i < len(split)-1 && !strings.HasPrefix(split[i+1], "--") {
			val = trimSpaceAndQuotes(split[i+1])
		}
		if arg != "" {
			m[arg] = val
		}
	}
	return m
}

func trimSpaceAndQuotes(s string) string {
	trimmed := strings.TrimSpace(s)
	trimmed = strings.TrimPrefix(trimmed, "\"")
	trimmed = strings.TrimPrefix(trimmed, "'")
	trimmed = strings.TrimSuffix(trimmed, "\"")
	trimmed = strings.TrimSuffix(trimmed, "'")
	return trimmed
}

func parseInt(s string, min int, max int) (int, error) {
	i64, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	i := int(i64)

	if i < min {
		return 0, fmt.Errorf("number must be greater than or equal to %d", min)
	}

	if i > max {
		return 0, fmt.Errorf("number must be less than or equal to %d", max)
	}
	return i, nil
}
