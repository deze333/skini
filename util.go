package skini

/*
Util -- helper functions.
*/

import (
	"fmt"
	"regexp"
    "unicode"
)

//------------------------------------------------------------
// Utility functions
//------------------------------------------------------------

// Camelcases and removes dots from key name.
// Name must not be empty.
// Example: server.www --> ServerWww
func toFieldName(s string) string {
    if s == "" {
        return ""
    }

	runes := []rune(s)
	out := []rune{unicode.ToUpper(runes[0])}
    wasDot := false

	for i := 1; i < len(runes); i++ {
		switch runes[i] {
		case rune('.'):
            wasDot = true
		default:
            if wasDot {
                out = append(out, unicode.ToUpper(runes[i]))
                wasDot = false
            } else {
                out = append(out, runes[i])
            }
		}
	}
    return string(out)
}

// Converts wildcard based filename pattern into regex
func wildcardRegex(pattern string) (re *regexp.Regexp, err error) {
	runes := []rune(pattern)
	expr := []rune{}
	expr = append(expr, []rune("^")...)
	for i := 0; i < len(runes); i++ {
		switch runes[i] {
		case rune('.'):
			expr = append(expr, []rune(`\.`)...)
		case rune('*'):
			expr = append(expr, []rune(`.*`)...)
		default:
			expr = append(expr, runes[i])
		}
	}
	expr = append(expr, []rune("$")...)

	re, err = regexp.Compile(string(expr))
	if err != nil {
		return re, fmt.Errorf("error compiling regex for pattern: %s", pattern)
	}
	return
}

