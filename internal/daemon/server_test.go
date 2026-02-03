package daemon

import (
	"bufio"
	"errors"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pxcli/internal/protocol"
	"pxcli/internal/testutil"
)

type stubHandler struct {
	response string
}

func (s stubHandler) Handle(request protocol.Request) string {
	return s.response
}

func TestServerRespondsSingleLine(t *testing.T) {
	socketPath := filepath.Join(testutil.TempDir(t), "pxcli.sock")
	server, err := NewServer(socketPath, stubHandler{response: "ok"})
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}
	done := startServer(t, server)
	t.Cleanup(func() {
		stopServer(t, server, done)
	})

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("unexpected error connecting to socket: %v", err)
	}
	defer conn.Close()

	if _, err := io.WriteString(conn, "clear\n"); err != nil {
		t.Fatalf("unexpected error writing request: %v", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("unexpected error reading response: %v", err)
	}
	if line != "ok\n" {
		t.Fatalf("expected response %q, got %q", "ok\n", line)
	}

	assertConnClosed(t, conn)
}

func TestServerInvalidRequestClosesConnection(t *testing.T) {
	socketPath := filepath.Join(testutil.TempDir(t), "pxcli.sock")
	server, err := NewServer(socketPath, stubHandler{response: "ok"})
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}
	done := startServer(t, server)
	t.Cleanup(func() {
		stopServer(t, server, done)
	})

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("unexpected error connecting to socket: %v", err)
	}
	defer conn.Close()

	if _, err := io.WriteString(conn, "   \n"); err != nil {
		t.Fatalf("unexpected error writing request: %v", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("unexpected error reading response: %v", err)
	}
	if !strings.HasPrefix(line, "err invalid_command ") {
		t.Fatalf("expected invalid_command error, got %q", line)
	}

	assertConnClosed(t, conn)
}

func startServer(t *testing.T, server *Server) <-chan error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- server.Serve()
	}()
	return done
}

func stopServer(t *testing.T, server *Server, done <-chan error) {
	t.Helper()
	_ = server.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("server did not shut down in time")
	}
}

func assertConnClosed(t *testing.T, conn net.Conn) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	var buf [1]byte
	n, err := conn.Read(buf[:])
	if n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after response, got n=%d err=%v", n, err)
	}
}
