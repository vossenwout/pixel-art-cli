package protocol

import "strings"

// Request represents a parsed protocol request line.
type Request struct {
	Command string
	Args    []string
}

// Error is a structured protocol error with a code and message.
type Error struct {
	Code    string
	Message string
}

func (e Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

// ParseLine parses a request line into a command and arguments.
func ParseLine(line string) (Request, error) {
	fields := fieldsASCII(line)
	if len(fields) == 0 {
		return Request{}, Error{Code: "invalid_command", Message: "command is required"}
	}
	return Request{Command: fields[0], Args: fields[1:]}, nil
}

// FormatOK formats a success response with an optional payload.
func FormatOK(payload string) string {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return "ok"
	}
	return "ok " + trimmed
}

// FormatError formats a failure response with a code and message.
func FormatError(code, message string) string {
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		trimmedCode = "error"
	}
	trimmedMessage := strings.TrimSpace(message)
	if trimmedMessage == "" {
		trimmedMessage = "unknown error"
	}
	return "err " + trimmedCode + " " + trimmedMessage
}

func fieldsASCII(s string) []string {
	if s == "" {
		return nil
	}
	fields := make([]string, 0, 4)
	for i := 0; i < len(s); {
		for i < len(s) && isASCIIWhitespace(s[i]) {
			i++
		}
		start := i
		for i < len(s) && !isASCIIWhitespace(s[i]) {
			i++
		}
		if start < i {
			fields = append(fields, s[start:i])
		}
	}
	return fields
}

func isASCIIWhitespace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}
