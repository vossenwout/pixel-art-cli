# Implementation tasks

Implementation tasks for building the specs outlined in the design doc: [prd](prd.md).

**Task Format**

## How to write a good task

- One task = one implementable unit of work.
- Numbering: use sequential integers (`T:1`, `T:2`, ...). Never reuse numbers.
- Title: verb + object + context (e.g. "Create a new canvas").
- Description: 1-3 sentences describing the feature/behavior to implement (no personas/benefits).
- Acceptance criteria must be testable/observable; cover happy path + key error case(s). Ofcourse also write tests and run them to verify this before you check it off.

## Task Template

### T:$NUMBER$: $TITLE$

**Description**
$DESCRIBE_THE_FEATURE_TO_IMPLEMENT_IN_1_3_SENTENCES.
$INCLUDE_KEY_BEHAVIOR_AND_CONSTRAINTS (inputs/outputs, CLI command/flags, defaults).

**Acceptance criteria:**

- [ ] Given $CONTEXT, when $ACTION, then $EXPECTED_RESULT
- [ ] Error: given $BAD_INPUT, when $ACTION, then $ERROR_BEHAVIOR
- [ ] Output: the CLI prints/returns $EXPECTED_OUTPUT (example included if needed)

## Example

### T:1: Create a new canvas

**Description**
Implement a `pixel-art new <width> <height>` command that creates a new canvas with the given dimensions.
The command validates inputs and prints a helpful error message on invalid usage.

**Acceptance criteria:**

- [ ] Given width=16 and height=16, when I run `pixel-art new 16 16`, then a 16x16 canvas is created
- [ ] Given a negative width, when I run `pixel-art new -1 16`, then the command exits non-zero and prints a helpful error
- [ ] Given no args, when I run `pixel-art new`, then the CLI prints usage/help for the command

## Tasks

### T:1: Scaffold Go module and Cobra CLI

**Description**
Create the Go module and the `pxcli` Cobra root command with a standard project layout (`cmd/pxcli`, `internal/...`).
Wire `--help` output, version plumbing (at least a `--version` flag), and ensure the binary builds and tests run locally.

**Acceptance criteria:**

- [x] Given a clean checkout, when I run `go build ./cmd/pxcli`, then the build succeeds and produces a runnable `pxcli` binary
- [x] Given an unknown command, when I run `pxcli does-not-exist`, then the CLI exits non-zero and prints command usage/help
- [x] Tests: when I run `go test ./...`, then it exits zero

### T:2: Define shared defaults and path configuration

**Description**
Define shared defaults for socket path (`/tmp/pxcli.sock`), PID file path (`/tmp/pxcli.pid`), default canvas size (`32x32`), default scale (`10`, reserved for windowed mode), and default headless mode (`true`).
Centralize these defaults in a small config package and ensure both the daemon and CLI resolve paths consistently.
Add a persistent `--socket` flag to all CLI subcommands and a matching flag for the daemon so the same binary can be pointed at a non-default socket. Add a `--headless` flag to `pxcli start` and `pxcli daemon` (default true).
Allow the daemon PID file path to be overridden via internal configuration (no user-facing flag required) so unit/integration tests can use temp paths and never touch `/tmp/pxcli.pid`.

**Acceptance criteria:**

- [x] Given no flags, when I run any client command, then it targets `/tmp/pxcli.sock` by default
- [x] Error: given `--socket` is an empty string, when I run any command, then the CLI exits non-zero and prints a helpful validation error
- [x] Given default config, when the daemon mode is resolved, then headless is true
- [x] Tests: tests can configure socket/PID paths to temp locations (no test reads/writes `/tmp/pxcli.pid`) and `go test ./...` passes

### T:3: Implement IPC protocol parsing and response formatting

**Description**
Implement a small protocol package that can parse a request line into `{command, args}` and can format responses as a single line.
Responses must follow: success `ok` or `ok <payload>` and failure `err <code> <message>`.
Parsing must ignore leading/trailing whitespace and treat 1+ ASCII whitespace characters as separators (i.e., `"set_pixel   1  2  red"` parses like `"set_pixel 1 2 red"`).

**Acceptance criteria:**

- [x] Given `set_pixel 10 10 #ff0000`, when it is parsed, then the command is `set_pixel` with args `["10","10","#ff0000"]`
- [x] Given `  get_pixel   1\t2  `, when it is parsed, then the command is `get_pixel` with args `["1","2"]`
- [x] Error: given an empty/whitespace-only request line, when it is parsed, then the daemon returns `err invalid_command <message>`
- [x] Tests: protocol parsing/formatting has unit tests and `go test ./...` passes

### T:4: Implement color parsing and canonical formatting

**Description**
Implement parsing for hex colors `#rgb`, `#rrggbb`, `#rrggbbaa` and named colors `red`, `blue`, `transparent` into RGBA.
Implement canonical formatting used by `get_pixel` responses as lowercase `#rrggbbaa`.

**Acceptance criteria:**

- [x] Given `#f00`, when it is parsed, then it becomes RGBA(255,0,0,255)
- [x] Error: given `#12` or `not-a-color`, when it is parsed, then the result is `err invalid_color <message>`
- [x] Output: given a fully transparent pixel, when it is formatted, then the string is `#00000000`
- [x] Tests: color parsing/formatting is unit tested and `go test ./...` passes

### T:5: Implement in-memory canvas model with bounds checking

**Description**
Implement a `Canvas` type that stores width/height and a pixel buffer of RGBA values.
Provide methods for `SetPixel(x,y,color)`, `GetPixel(x,y)`, and `Clear(color)` with explicit bounds validation.

**Acceptance criteria:**

- [x] Given a 4x4 canvas, when I set pixel (1,2) to red and then get (1,2), then the returned color is red
- [x] Error: given x=-1 or x>=width (or similarly for y), when I call `GetPixel`/`SetPixel`, then it returns an `out_of_bounds` error
- [x] Tests: canvas get/set/clear and bounds behavior are unit tested and `go test ./...` passes

### T:6: Implement `fill_rect` drawing operation

**Description**
Implement `FillRect(x,y,w,h,color)` that fills pixels in the rectangle `[x, x+w)` by `[y, y+h)`.
Validate that `w>0`, `h>0`, and the full rectangle lies within the canvas; otherwise return a structured error.

**Acceptance criteria:**

- [x] Given a 8x8 canvas, when I fill_rect 2 3 3 2 with blue, then pixels (2..4,3..4) are blue and others are unchanged
- [x] Error: given `w<=0` or a rectangle that exceeds canvas bounds, when `fill_rect` is executed, then the daemon returns `err out_of_bounds <message>` (or `err invalid_args <message>` for non-positive sizes)
- [x] Tests: fill_rect behavior and error cases are unit tested and `go test ./...` passes

### T:7: Implement `line` drawing operation

**Description**
Implement `Line(x1,y1,x2,y2,color)` using a deterministic integer line algorithm (e.g., Bresenham) including both endpoints.
Validate both endpoints are within bounds; if not, return `out_of_bounds`.

**Acceptance criteria:**

- [x] Given a 8x8 canvas, when I draw `line 0 0 3 0 red`, then pixels (0,0), (1,0), (2,0), (3,0) are red
- [x] Error: given x2 is outside the canvas, when `line` is executed, then the daemon returns `err out_of_bounds <message>`
- [x] Tests: horizontal/vertical/diagonal lines are unit tested and `go test ./...` passes

### T:8: Implement undo/redo history for mutating operations

**Description**
Implement snapshot-based undo/redo history in the daemon for all mutating commands (`set_pixel`, `fill_rect`, `line`, `clear`).
Undo restores the previous canvas state; redo reapplies; any new mutating operation clears the redo stack.

**Acceptance criteria:**

- [x] Given a canvas where a pixel was changed, when I run `undo`, then the canvas returns to the previous state
- [x] Error: given no history is available, when I run `undo` (or `redo`), then the daemon returns `err no_history <message>`
- [x] Tests: undo/redo stack behavior (including redo clearing) is unit tested and `go test ./...` passes

### T:9: Implement PNG export from canvas state

**Description**
Implement exporting the current canvas to a PNG file using Go's `image` and `image/png` packages.
The export command writes exactly the canvas dimensions (no scaling) and returns an I/O error code on failure.

**Acceptance criteria:**

- [x] Given a 2x2 canvas with known colors, when I export to a temp file and decode the PNG, then the decoded pixels match the canvas exactly
- [x] Error: given a path that cannot be created/written, when I run export, then the daemon returns `err io <message>`
- [x] Tests: PNG export has unit tests that decode and verify output and `go test ./...` passes

### T:10: Implement daemon command handler (request -> canvas operations)

**Description**
Implement a command handler that maps protocol requests to canvas operations and returns exactly one response line.
Support: `set_pixel`, `get_pixel`, `fill_rect`, `line`, `clear [color]`, `export <filename>`, `undo`, `redo`, `stop`.

**Acceptance criteria:**

- [x] Given `get_pixel 1 2` on a known canvas, when the handler executes it, then it returns `ok #rrggbbaa` matching the stored pixel
- [x] Error: given a command with the wrong arg count (e.g., `set_pixel 1 2`), when it is handled, then the response is `err invalid_args <message>`
- [x] Error: given an unknown command (e.g., `nope`), when it is handled, then the response is `err invalid_command <message>`
- [x] Output: given a successful mutating command, when handled, then the response is exactly `ok` (single line)
- [x] Given `clear` with no args, when handled, then the canvas is cleared to transparent and the response is `ok`
- [x] Given `stop`, when handled, then it responds `ok` and triggers daemon shutdown
- [x] Tests: command handler behavior is unit tested (including error codes) and `go test ./...` passes

### T:11: Implement Unix domain socket server (one request per connection)

**Description**
Implement a Unix domain socket server that listens on the configured socket path and handles one request line per connection.
The server must process connections sequentially (one at a time) and block additional clients until the active connection closes.

**Acceptance criteria:**

- [x] Given the daemon is listening, when a client connects, sends one newline-terminated request, and reads, then it receives exactly one newline-terminated response line
- [x] Error: given a client connects and sends an invalid request, when handled, then the server responds with `err <code> <message>` and closes the connection
- [x] Tests: socket server has integration tests using a temp socket path and `go test ./...` passes

### T:12: Implement PID file + stale socket handling

**Description**
Implement daemon startup checks using `/tmp/pxcli.pid` and the socket file: if the PID is alive, refuse to start with a clear error; if stale, remove PID and socket before binding.
If the PID file is missing but the socket path already exists, attempt to connect: if a connection succeeds, treat it as "already running"; if it fails, treat it as stale and remove the socket before binding.
Write the daemon PID on startup and remove PID + socket on clean shutdown.

**Acceptance criteria:**

- [x] Given a stale PID file for a non-existent process, when the daemon starts, then it removes the stale PID + socket and successfully binds the socket
- [x] Error: given an active daemon PID is recorded, when a second daemon starts, then it exits non-zero with `err daemon_already_running <message>`
- [x] Given no PID file but an existing stale socket file, when the daemon starts, then it removes the socket and successfully binds
- [x] Tests: PID liveness logic is unit tested via an injectable liveness check and `go test ./...` passes

### T:13: Implement graceful shutdown (stop command, signals, cleanup)

**Description**
Implement graceful shutdown paths: `stop` command triggers shutdown, SIGINT/SIGTERM trigger shutdown, and all paths remove socket + PID file.
Ensure shutdown completes even if no clients are connected.

**Acceptance criteria:**

- [x] Given the daemon is running, when it receives a `stop` request, then it responds `ok`, exits, and removes `/tmp/pxcli.sock` and `/tmp/pxcli.pid`
- [x] Given the daemon is running, when it receives SIGINT, then it exits and removes `/tmp/pxcli.sock` and `/tmp/pxcli.pid`
- [x] Tests: shutdown coordination (headless, no GUI) is covered by tests and `go test ./...` passes

### T:14: Implement headless daemon runtime (no GUI)

**Description**
Implement a headless daemon runtime that wires the socket server, command handler, and shutdown coordination without initializing any renderer or window.
It must run with no display environment and rely on the in-memory canvas plus `export`/`get_pixel` for observability.

**Acceptance criteria:**

- [x] Given the headless runtime is started, when a client sends a valid command, then it responds `ok` and updates the canvas state
- [x] Given `DISPLAY` is unset, when the headless runtime starts, then it still serves requests normally
- [x] Tests: headless runtime start/stop is covered in tests and `go test ./...` passes

### T:15: Implement `pxcli daemon` subcommand (hidden internal entrypoint)

**Description**
Add a hidden `pxcli daemon` subcommand that starts the daemon process: initializes canvas state, starts the socket server, and runs the headless runtime.
Support flags: `--size 32x32`, `--scale 10` (reserved for future windowed mode), `--headless` (default true), `--socket <path>` (and use the default PID path unless otherwise configured internally).

**Acceptance criteria:**

- [x] Given `pxcli daemon --size 32x32 --headless`, when it starts, then it listens on the socket and serves requests
- [x] Error: given an invalid size string (e.g., `--size 0x10` or `--size 10`), when it starts, then it exits non-zero and prints a helpful validation error
- [x] Error: given an invalid scale (e.g., `--scale 0` or `--scale -1`), when it starts, then it exits non-zero and prints a helpful validation error
- [x] Tests: size parsing/validation is unit tested and `go test ./...` passes

### T:16: Implement CLI client transport (connect/send/receive)

**Description**
Implement a client package used by CLI subcommands to connect to the Unix socket, send a single request line, and read the single-line response.
Return structured errors for connection failures and for `err <code> <message>` responses.
Use reasonable deadlines (connect + read/write) so CLI commands do not hang indefinitely if the daemon is unresponsive.

**Acceptance criteria:**

- [x] Given a test server that replies `ok`, when the client sends `clear`, then it returns success and exposes the response payload (if any)
- [x] Error: given the socket path does not exist, when the client sends a command, then it returns an error that the CLI prints as `err daemon_not_running <message>` and exits non-zero
- [x] Error: given a server accepts a connection but never replies, when the client sends a command, then it returns a timeout error (within a bounded duration)
- [x] Tests: client transport has unit tests using a temp socket and `go test ./...` passes

### T:17: Implement `pxcli start` (spawn daemon in background)

**Description**
Implement `pxcli start [--size 32x32] [--scale 10] [--headless] [--socket <path>]` which spawns the daemon as a detached background process (headless by default).
Print the daemon PID and exit; poll until the socket is accepting connections (or time out with a clear error) before returning success. `--scale` is accepted for future windowed mode but has no effect in headless mode.

**Acceptance criteria:**

- [ ] Given no daemon is running, when I run `pxcli start`, then it exits zero, prints the daemon PID, and the daemon is ready to accept requests (headless)
- [ ] Error: given `--scale 0` (or `--scale -1`), when I run `pxcli start`, then it exits non-zero and prints a helpful validation error
- [ ] Error: given a daemon is already running, when I run `pxcli start`, then it exits non-zero and prints `err daemon_already_running <message>`
- [ ] Output: given `pxcli start --size 8x8 --headless`, when it succeeds, then subsequent client commands operate on an 8x8 canvas
- [ ] Tests: start command argument validation and daemon argv construction are unit tested and `go test ./...` passes
 - [x] Given no daemon is running, when I run `pxcli start`, then it exits zero, prints the daemon PID, and the daemon is ready to accept requests (headless)
 - [x] Error: given `--scale 0` (or `--scale -1`), when I run `pxcli start`, then it exits non-zero and prints a helpful validation error
 - [x] Error: given a daemon is already running, when I run `pxcli start`, then it exits non-zero and prints `err daemon_already_running <message>`
 - [x] Output: given `pxcli start --size 8x8 --headless`, when it succeeds, then subsequent client commands operate on an 8x8 canvas
 - [x] Tests: start command argument validation and daemon argv construction are unit tested and `go test ./...` passes

### T:18: Implement `pxcli stop` (request shutdown and wait)

**Description**
Implement `pxcli stop` that sends a `stop` request to the daemon and waits for clean shutdown.
If the daemon is not reachable, return a clear non-zero error.

**Acceptance criteria:**

- [x] Given a daemon is running, when I run `pxcli stop`, then it prints `ok` and the socket/PID files are removed
- [x] Error: given no daemon is running, when I run `pxcli stop`, then it exits non-zero and prints `err daemon_not_running <message>`
- [x] Tests: stop behavior is exercised via integration tests (headless, no GUI) and `go test ./...` passes

### T:19: Implement CLI drawing commands (set_pixel, fill_rect, line, clear)

**Description**
Add Cobra subcommands that validate arguments and send protocol requests: `set_pixel <x> <y> <color>`, `fill_rect <x> <y> <w> <h> <color>`, `line <x1> <y1> <x2> <y2> <color>`, `clear [color]`.
Commands must print the daemon response line and exit non-zero on `err` responses.

**Acceptance criteria:**

- [x] Given a running daemon, when I run `pxcli set_pixel 10 10 #ff0000`, then it prints `ok` and `pxcli get_pixel 10 10` returns `ok #ff0000ff`
- [x] Error: given a bad arg (e.g., `pxcli fill_rect 0 0 -1 2 red`), when I run it, then it exits non-zero and prints `err invalid_args <message>`
- [x] Tests: CLI arg validation and request formatting are unit tested (via a mocked client) and `go test ./...` passes

### T:20: Implement CLI utility commands (get_pixel, export)

**Description**
Add `pxcli get_pixel <x> <y>` and `pxcli export <filename.png>` commands.
`get_pixel` prints the daemon response (including the returned `#rrggbbaa` color) and `export` writes the file via the daemon.
For predictable behavior with a background daemon, the CLI must resolve `export` paths to absolute paths (relative to the CLI's current working directory) before sending the request to the daemon.

**Acceptance criteria:**

- [x] Given a running daemon with a known pixel, when I run `pxcli get_pixel 0 0`, then it prints `ok #rrggbbaa` with the correct color
- [x] Given I run `pxcli export out.png` from a directory, when it succeeds, then `out.png` is created in that directory (even though the daemon is a separate process)
- [x] Error: given an unwritable filename path, when I run `pxcli export /root/out.png`, then it exits non-zero and prints `err io <message>`
- [x] Tests: CLI behaviors for get_pixel/export are covered and `go test ./...` passes

### T:21: Implement CLI history commands (undo, redo)

**Description**
Add `pxcli undo` and `pxcli redo` commands that forward the request to the daemon and print the single-line response.
Ensure exit codes match daemon success/error.

**Acceptance criteria:**

- [ ] Given at least one prior mutating command, when I run `pxcli undo`, then it prints `ok` and a subsequent `pxcli get_pixel` reflects the previous state
- [ ] Error: given no redo is available, when I run `pxcli redo`, then it exits non-zero and prints `err no_history <message>`
- [ ] Tests: CLI undo/redo request wiring is unit tested and `go test ./...` passes

### T:22: Add end-to-end protocol integration tests (daemon core without UI)

**Description**
Create an integration test harness that starts the daemon core (socket server + command handler) in headless mode (no GUI) and exercises the full request/response protocol over a real Unix socket.
Cover at least: set_pixel/get_pixel, fill_rect, line, clear, export (to temp dir), undo/redo, and stop.

**Acceptance criteria:**

- [ ] Given the test harness, when it sends the documented command set over the socket, then responses match the `ok`/`err` protocol and state changes are correct
- [ ] Error: given an out-of-bounds coordinate, when the test sends `set_pixel -1 0 red`, then it receives `err out_of_bounds <message>`
- [ ] Tests: these tests use temp socket/PID paths and never read/write `/tmp/pxcli.sock` or `/tmp/pxcli.pid`
- [ ] Tests: `go test ./...` runs these integration tests and exits zero

### T:23: Document usage and protocol expectations

**Description**
Write user-facing documentation covering headless mode, how to start/stop the daemon, available drawing commands, color formats, and example request/response lines.
Include troubleshooting notes for stale PID/socket cleanup and how to override `--socket`.
Document that headless mode is the default for development and that windowed GUI support will be added later.
Document how `pxcli export <path>` is resolved (relative paths are interpreted from the CLI's current working directory).

**Acceptance criteria:**

- [ ] Given the repository README/docs, when a user follows the steps, then they can start the daemon in headless mode, draw a pixel, query it, export a PNG, and stop the daemon
- [ ] Given a stale PID/socket scenario, when the user follows the troubleshooting steps, then they can recover without manual file deletion
- [ ] Tests: documentation changes do not break the build and `go test ./...` passes

### T:24: Ensure daemon concurrency is race-free

**Description**
Ensure shared daemon state (canvas buffer, history, shutdown flags) is safe when accessed by the socket server and shutdown coordination concurrently (headless mode).
Add/adjust tests and synchronization so the project passes the Go race detector.

**Acceptance criteria:**

- [ ] Given the full test suite, when I run `go test -race ./...`, then it exits zero and reports no data races
- [ ] Given a running daemon, when I rapidly issue drawing commands while the daemon is processing requests, then it does not crash or deadlock
- [ ] Tests: `go test ./...` still passes

## Future (GUI) work (not in scope for headless development)

- Add Ebiten renderer and window lifecycle behavior (`--scale`, close-to-exit)
- Wire the renderer into the daemon runtime and reconcile concurrent access with the render loop
