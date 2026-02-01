package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

const (
	defaultDialTimeout  = 2 * time.Second
	defaultWriteTimeout = 2 * time.Second
	defaultReadTimeout  = 2 * time.Second
)

// Error represents a structured client error with a code and message.
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

// Response represents a parsed daemon response line.
type Response struct {
	Raw     string
	Payload string
}

// Client sends protocol requests to the daemon socket.
type Client struct {
	socketPath   string
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// Option configures the client.
type Option func(*Client)

// New creates a client for the provided socket path.
func New(socketPath string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(socketPath) == "" {
		return nil, fmt.Errorf("socket path must not be empty")
	}
	client := &Client{
		socketPath:   socketPath,
		dialTimeout:  defaultDialTimeout,
		readTimeout:  defaultReadTimeout,
		writeTimeout: defaultWriteTimeout,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}
	return client, nil
}

// WithDialTimeout overrides the default dial timeout.
func WithDialTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.dialTimeout = timeout
		}
	}
}

// WithReadTimeout overrides the default read timeout.
func WithReadTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.readTimeout = timeout
		}
	}
}

// WithWriteTimeout overrides the default write timeout.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.writeTimeout = timeout
		}
	}
}

// Send sends a single request line and returns the parsed response.
func (c *Client) Send(request string) (Response, error) {
	if c == nil {
		return Response{}, Error{Code: "invalid_client", Message: "client is nil"}
	}
	trimmed := strings.TrimRight(request, "\r\n")
	if strings.TrimSpace(trimmed) == "" {
		return Response{}, Error{Code: "invalid_request", Message: "request is required"}
	}

	dialer := net.Dialer{Timeout: c.dialTimeout}
	conn, err := dialer.Dial("unix", c.socketPath)
	if err != nil {
		return Response{}, classifyDialError(err)
	}
	defer conn.Close()

	if err := conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return Response{}, Error{Code: "connection_failed", Message: err.Error()}
	}
	if _, err := io.WriteString(conn, trimmed+"\n"); err != nil {
		return Response{}, Error{Code: "io", Message: err.Error()}
	}

	if err := conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
		return Response{}, Error{Code: "connection_failed", Message: err.Error()}
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		if isTimeout(err) {
			return Response{}, Error{Code: "timeout", Message: "timed out waiting for response"}
		}
		return Response{}, Error{Code: "io", Message: err.Error()}
	}
	line = strings.TrimRight(line, "\r\n")
	return parseResponse(line)
}

func parseResponse(line string) (Response, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return Response{}, Error{Code: "invalid_response", Message: "empty response"}
	}
	if hasTokenPrefix(trimmed, "ok") {
		payload := strings.TrimSpace(trimmed[len("ok"):])
		return Response{Raw: trimmed, Payload: payload}, nil
	}
	if hasTokenPrefix(trimmed, "err") {
		rest := strings.TrimSpace(trimmed[len("err"):])
		if rest == "" {
			return Response{}, Error{Code: "error", Message: "unknown error"}
		}
		parts := strings.SplitN(rest, " ", 2)
		code := parts[0]
		message := "unknown error"
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			message = parts[1]
		}
		return Response{Raw: trimmed}, Error{Code: code, Message: message}
	}
	return Response{}, Error{Code: "invalid_response", Message: fmt.Sprintf("unexpected response %q", trimmed)}
}

func hasTokenPrefix(value, token string) bool {
	if value == token {
		return true
	}
	return strings.HasPrefix(value, token+" ")
}

func classifyDialError(err error) error {
	if isTimeout(err) {
		return Error{Code: "timeout", Message: err.Error()}
	}
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT) || errors.Is(err, syscall.ECONNREFUSED) {
		return Error{Code: "daemon_not_running", Message: err.Error()}
	}
	return Error{Code: "connection_failed", Message: err.Error()}
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
