package daemon

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"

	"pxcli/internal/protocol"
)

// RequestHandler handles a parsed protocol request.
type RequestHandler interface {
	Handle(request protocol.Request) string
}

// Server listens on a Unix socket and handles one request per connection.
type Server struct {
	listener net.Listener
	handler  RequestHandler
}

// NewServer creates a server listening on the provided Unix socket path.
func NewServer(socketPath string, handler RequestHandler) (*Server, error) {
	if handler == nil {
		return nil, errors.New("handler must not be nil")
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	return &Server{listener: listener, handler: handler}, nil
}

// Serve accepts connections sequentially until the listener is closed.
func (s *Server) Serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		s.handleConn(conn)
	}
}

// Close shuts down the server listener.
func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		writeResponse(conn, protocol.FormatError("invalid_command", "unable to read request"))
		return
	}
	line = strings.TrimRight(line, "\r\n")
	request, err := protocol.ParseLine(line)
	if err != nil {
		writeResponse(conn, formatProtocolError(err))
		return
	}
	response := s.handler.Handle(request)
	writeResponse(conn, response)
}

func writeResponse(conn net.Conn, response string) {
	_, _ = io.WriteString(conn, response+"\n")
}

func formatProtocolError(err error) string {
	var perr protocol.Error
	if errors.As(err, &perr) {
		return protocol.FormatError(perr.Code, perr.Message)
	}
	return protocol.FormatError("invalid_command", err.Error())
}
