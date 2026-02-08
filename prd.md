# CLI Tool For Drawing Pixel Art

## Description

A local pixel art tool designed to be controlled programmatically by AI coding agents (like Claude Code). It consists of a long-running daemon that runs with a windowed GUI by default when built with the `ebiten` tag and accepts drawing commands over a Unix domain socket (request/response), plus a CLI that sends commands to the daemon. Headless mode is available via `--headless` for CI/containers.

## Technology choices

Language: Go
Graphics (windowed mode): Ebiten (simple 2D library, build-tagged)
Canvas buffer (headless): Go image.RGBA
CLI parsing: Cobra
IPC: Unix domain socket
Image export: Go's image/png package

## Architecture

┌─────────────────┐         ┌─────────────────────────────┐
│   CLI (pxcli)   │<------->│  Unix Domain Socket         │
│                 │ connect │  /tmp/pxcli.sock            │
└─────────────────┘         └──────────────┬──────────────┘
                                            │ request/response
                                            v
                            ┌─────────────────────────────┐
                            │  Daemon core                │
                            │  - Canvas state in memory   │
                            │  - Parses commands          │
                            │  - Updates pixel buffer     │
                            └──────────────┬──────────────┘
                                           │ (optional)
                                           v
                            ┌─────────────────────────────┐
                            │  Renderer (windowed, optional)
                            │  - Ebiten window            │
                            │  - Displays canvas scaled   │
                            └─────────────────────────────┘

Headless mode omits the renderer/window; `export` and `get_pixel` provide output for agents.

## Components

1. Daemon (long-running process)

Runs in windowed mode by default when built with `ebiten`; headless mode is available via `--headless`
Listens on a Unix domain socket
Holds canvas state in memory (2D array of RGBA values)
Accepts connections, reads a command, executes it, writes a response
Runs until stop command or signal; in windowed mode, closing the window also terminates the daemon

1. CLI (thin client)

Single binary with subcommands
pxcli start — spawns daemon in background
pxcli stop — sends stop command to daemon
pxcli <command> [args] — sends command to daemon and prints response / errors

1. Unix Domain Socket

The daemon listens on a single socket path (or configurable)
Simple line-based request/response protocol (one command per request)
Example request: set_pixel 10 10 #ff0000
Example response: ok

1. Windowed renderer (Ebiten)

Ebiten window that displays the canvas scaled up so pixels are visible
Only used in windowed mode; headless mode runs without any display
Compiled behind an `ebiten` build tag so headless builds do not require GUI deps

## Daemon Lifecycle

**Starting the daemon:**
`pxcli start` spawns the daemon as a background process using fork-and-detach. The CLI prints the daemon PID and exits. The daemon writes its PID to `/tmp/pxcli.pid`.

**Stopping the daemon:**

- `pxcli stop` sends a stop command and waits for clean shutdown
- Headless mode runs until `stop` or a signal; windowed mode also exits when the window is closed
- SIGTERM/SIGINT trigger graceful shutdown

**Stale socket handling:**
Before binding `/tmp/pxcli.sock`, the daemon checks if a PID file exists. If the recorded process is dead (stale), it removes the socket and PID file. If alive, it exits with an error ("daemon already running").

## Headless mode

- Headless is opt-in for development and CI; pass `--headless`.
- No window is created; the daemon can run with no display environment.
- `export` and `get_pixel` are the primary ways to observe output.
- `--scale` is reserved for windowed mode and has no effect in headless mode.
- EC2 development must remain headless; windowed verification is local only.

## Windowed mode (Ebiten)

- Windowed mode is the default when built with the `ebiten` build tag.
- Uses `--scale` to size the window (`canvas_size * scale`) with pixel-perfect rendering.
- Closing the window exits the daemon; `stop` should close the window cleanly.
- If the binary is built without the `ebiten` tag, starting without `--headless` returns a clear error.

## Canvas targeting (future)

Initial version assumes a single active canvas in the daemon.

If/when multiple canvases are needed within a single daemon, commands will accept an optional `canvas_id` (with a default canvas selected when omitted).

## IPC protocol

The protocol is ASCII text, one request per line. Each request receives exactly one response line.

Responses:

- Success: `ok` or `ok <payload>`
- Failure: `err <code> <message>`

Examples:

- `set_pixel 10 10 #ff0000` -> `ok`
- `get_pixel 10 10` -> `ok #ff0000ff`
- `set_pixel -1 10 #ff0000` -> `err out_of_bounds x must be >= 0`

## Connection Model

The daemon accepts one CLI connection at a time. When a connection is active, subsequent connection attempts block until the current one closes. This simplifies implementation and is sufficient for sequential AI agent commands.

## Expected API

**Daemon control**
pxcli start [--size 32x32] [--scale 10] [--headless]   # start daemon (defaults: 32x32 canvas, windowed; scale used only in windowed mode)
pxcli stop                                # shutdown the target daemon

**Drawing**
pxcli set_pixel <x> <y> <color>            # set single pixel
pxcli fill_rect <x> <y> <w> <h> <color>    # filled rectangle
pxcli line <x1> <y1> <x2> <y2> <color>     # draw line
pxcli clear [color]                         # clear canvas (default: transparent)

**Utility**
pxcli export <filename.png>                # save canvas to file
pxcli get_pixel <x> <y>                    # return color at position (for AI to query)
pxcli undo                                  # undo last operation
pxcli redo                                  # redo

**Color format**

Hex: #ff0000, #f00, #ff000080 (with alpha)
Or named: red, blue, transparent
