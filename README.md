# pxcli

pxcli is a headless pixel-art daemon and CLI that communicates over a Unix domain socket.
It is designed for programmatic control: a long-running daemon keeps the canvas in memory,
while the CLI sends one-line commands and prints one-line responses.

## Quick start (headless)

Build the CLI:

```bash
go build ./cmd/pxcli
```

Start the daemon in the background (defaults to 32x32, headless):

```bash
./pxcli start
```

The command prints the daemon PID when it is ready.

Draw a pixel and read it back:

```bash
./pxcli set_pixel 1 1 #ff0000
./pxcli get_pixel 1 1
```

Export the canvas to a PNG in your current directory and stop the daemon:

```bash
./pxcli export out.png
./pxcli stop
```

## Commands

Daemon lifecycle:

- `pxcli start [--size 32x32] [--scale 10] [--headless] [--socket <path>]`
- `pxcli stop [--socket <path>]`

Drawing:

- `pxcli set_pixel <x> <y> <color>`
- `pxcli fill_rect <x> <y> <w> <h> <color>`
- `pxcli line <x1> <y1> <x2> <y2> <color>`
- `pxcli clear [color]`

Utility:

- `pxcli get_pixel <x> <y>`
- `pxcli export <filename.png>`
- `pxcli undo`
- `pxcli redo`

Notes:

- Headless mode is the default; windowed mode is build-tagged (`ebiten`) for local use.
- `--scale` is used only in windowed mode and has no effect in headless mode.
- EC2 development is headless-only; GUI verification is done locally.
- The hidden `pxcli daemon` subcommand is the internal entrypoint used by `pxcli start`.

## Defaults and configuration

Defaults:

- Socket: `/tmp/pxcli.sock`
- PID file: `/tmp/pxcli.pid`
- Canvas: `32x32`
- Headless: `true`

To target a non-default socket, pass `--socket` to every command (start, stop, and all drawing/utility commands):

```bash
./pxcli start --socket /tmp/alt.sock
./pxcli set_pixel 0 0 red --socket /tmp/alt.sock
./pxcli stop --socket /tmp/alt.sock
```

## Protocol summary

The daemon accepts one ASCII request line per connection and responds with one line:

- Success: `ok` or `ok <payload>`
- Failure: `err <code> <message>`

Examples:

- `set_pixel 10 10 #ff0000` -> `ok`
- `get_pixel 10 10` -> `ok #ff0000ff`
- `set_pixel -1 10 #ff0000` -> `err out_of_bounds x must be >= 0`

Connection model:

- The daemon handles one connection at a time; additional clients block until the active connection closes.
- Each connection is exactly one request and one response line.

## Color formats

Accepted input formats:

- Hex: `#rgb`, `#rrggbb`, `#rrggbbaa`
- Named: `red`, `blue`, `transparent`

`get_pixel` returns canonical lowercase `#rrggbbaa`.

## Export path behavior

`pxcli export <filename.png>` resolves the filename to an absolute path before sending the request.
That means relative paths are interpreted from the CLI's current working directory,
not from where the daemon is running.

## Troubleshooting stale PID/socket files

If the daemon crashes or is killed, `/tmp/pxcli.pid` or `/tmp/pxcli.sock` may remain:

- Re-run `pxcli start` (or `pxcli daemon`) to automatically remove stale PID/socket files.
- If you see `err daemon_already_running`, a live daemon is likely running on that socket.
  Use `pxcli stop --socket <path>` to shut it down, or point your CLI to the correct socket.

Manual deletion should not be necessary for normal recovery.

## Headless-first development

Headless mode is the default for development and CI. Windowed GUI support (Ebiten) is available only
when you build with the `ebiten` tag, and it should be verified locally (not inside the headless container).

## Windowed GUI (Ebiten)

Windowed mode uses the `ebiten` build tag to avoid GUI dependencies in CI/Development within docker container.
Local usage:

```bash
go build -tags=ebiten ./cmd/pxcli
./pxcli start --headless=false --scale 10
```

Notes:

- The headless container/CI environment cannot open a window; build and run windowed mode locally.
- `--scale` controls the pixel-perfect window size (canvas size × scale).
