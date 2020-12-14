package cue

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

// parseCommand retrive string line and parses it with the following algorythm:
// * first word in the line is command name (cmd return value)
// * all rest words are command's parameters
// * if parameter includes more than one word it should be wrapped with ' or "
func parseCommand(line string) (cmd string, params []string, err error) {
	line = strings.TrimSpace(line)
	params = make([]string, 0)

	// Find cmd.
	i := strings.IndexFunc(line, unicode.IsSpace)
	if i < 0 { // We have only command without any parameters.
		cmd = line
		return
	}
	cmd = line[:i]
	line = strings.TrimSpace(line[i:])

	// Split parameters.
	l := len(line)
	var quotedChar byte = 0
	param := bytes.NewBufferString("")
	for i = 0; i < l; i++ {
		c := line[i]

		if quotedChar == 0 { // We are not in quote mode now, so we can enter into.
			if isQuoteChar(c) {
				// Quote can be started only at the beginnig of the parameter,
				// but not in the middle.
				if param.Len() != 0 {
					err = errors.New("unexpected quotation character")
					return
				}
				quotedChar = c
			} else if unicode.IsSpace(rune(c)) {
				// In not quote mode space starts new parameter.
				// But don't save empty parameters.
				if param.Len() != 0 {
					params = append(params, param.String())
					param = bytes.NewBufferString("")
				}
			} else {
				if c == '\\' { // Escape sequence in the text.
					if i+1 >= l {
						err = errors.New("unfinished escape sequence")
						return
					}

					s, e := parseEscapeSequence(line[i : i+2])
					if e != nil {
						param.WriteByte(c)
					} else {
						param.WriteByte(s)
						i++
					}
				} else {
					param.WriteByte(c)
				}
			}
		} else {
			if c == quotedChar { // Close quote.
				quotedChar = 0
			} else {
				if c == '\\' { // Escape sequence in the text.
					if i+1 >= l {
						err = errors.New("unfinished escape sequence")
						return
					}

					s, e := parseEscapeSequence(line[i : i+2])
					if e != nil {
						param.WriteByte(c)
					} else {
						param.WriteByte(s)
						i++
					}
				} else {
					param.WriteByte(c)
				}
			}
		}
	}

	params = append(params, param.String())

	return
}

// parseEscapeSequence returns escape character by it's string "source code" equivalent.
func parseEscapeSequence(seq string) (char byte, err error) {
	var m = map[string]byte{
		"\\\"": '"',
		"\\'":  '\'',
		"\\\\": '\\',
		"\\n":  '\n',
		"\\t":  '\t',
	}

	char, ok := m[seq]
	if !ok {
		err = fmt.Errorf("usupported escape sequence '%s'", seq)
	}

	return
}

// isQuoteChar returns true if given char is string quoted char:
// " or '.
func isQuoteChar(char byte) bool {
	return char == '\'' || char == '"'
}

// parserTime parses time string and returns separate values.
// Input string format: mm:ss:ff
func parseTime(length string) (min int, sec int, frames int, err error) {
	parts := strings.Split(length, ":")
	if len(parts) != 3 {
		err = errors.New("illegal time format, mm:ss:ff should be")
		return
	}

	min, err = strconv.Atoi(parts[0])
	if err != nil {
		err = errors.Wrap(err, "failed to parse minutes")
		return
	}

	sec, err = strconv.Atoi(parts[1])
	if err != nil {
		err = errors.Wrap(err, "failed to parse seconds")
		return
	}
	if sec > 59 {
		err = errors.New("failed to parse seconds, seconds value can't be more than 59")
		return
	}

	frames, err = strconv.Atoi(parts[2])
	if err != nil {
		err = errors.Wrap(err, "failed to parse frames value")
		return
	}
	if frames > framesPerSecond-1 {
		err = fmt.Errorf("failed to parse frames, frames value can't be more than %d", framesPerSecond-1)
		return
	}

	return
}
