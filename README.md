# pxcli

pxcli is a pixel-art daemon and CLI that communicates over a Unix domain socket.
It runs a long-lived daemon that keeps the canvas in memory while the CLI sends one-line commands
and prints one-line responses. By default it runs with a windowed GUI when built with `-tags=ebiten`;
pass `--headless` to run without a GUI.

## Quick start (windowed)

Build with GUI support:

```bash
go build -tags=ebiten ./cmd/pxcli
```

Start the daemon (defaults to 32x32, windowed):

```bash
./pxcli start
```

Draw, export, and stop:

```bash
./pxcli set_pixel 1 1 #ff0000
./pxcli export out.png
./pxcli stop
```

## Quick start (headless)

Build without GUI dependencies:

```bash
go build ./cmd/pxcli
```

Start the daemon in headless mode:

```bash
./pxcli start --headless
```

## CLI API

All commands accept the persistent `--socket <path>` flag to target a non-default socket.

Daemon lifecycle:

- `pxcli start [--size 32x32] [--scale 10] [--headless] [--socket <path>]`
- `pxcli stop [--socket <path>]`
- `pxcli daemon [--size 32x32] [--scale 10] [--headless] [--socket <path>]` (hidden, used by `start`)

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

Flags:

- `--size` expects `WxH` (example: `32x32`).
- `--scale` is used only in windowed mode and has no effect in headless mode.
- `--headless` runs without a GUI.

`pxcli start` prints the daemon PID when the socket is ready.

## Socket API (protocol)

The daemon accepts one ASCII request line per connection and responds with one line:

- Success: `ok` or `ok <payload>`
- Failure: `err <code> <message>`

Each connection is exactly one request and one response. Additional clients block until the active
connection closes.

Commands:

- `set_pixel <x> <y> <color>` -> `ok`
- `get_pixel <x> <y>` -> `ok #rrggbbaa`
- `fill_rect <x> <y> <w> <h> <color>` -> `ok`
- `line <x1> <y1> <x2> <y2> <color>` -> `ok`
- `clear [color]` -> `ok`
- `export <filename.png>` -> `ok`
- `undo` -> `ok`
- `redo` -> `ok`
- `stop` -> `ok`

Examples:

```
set_pixel 10 10 #ff0000
ok
```

```
get_pixel 10 10
ok #ff0000ff
```

Common error codes:

- `invalid_command` unknown command
- `invalid_args` wrong argument count or type
- `invalid_color` unsupported color format
- `out_of_bounds` coordinate outside canvas
- `no_history` undo/redo with empty history
- `io` export file error

## Color formats

Accepted input formats:

- Hex: `#rgb`, `#rrggbb`, `#rrggbbaa`
- Named: `black`, `white`, `red`, `green`, `blue`, `yellow`, `orange`, `purple`, `cyan`, `magenta`, `gray`, `grey`, `transparent`

`get_pixel` returns canonical lowercase `#rrggbbaa`.

## Defaults and configuration

Defaults:

- Socket: `/tmp/pxcli.sock`
- PID file: `/tmp/pxcli.pid`
- Canvas: `32x32`
- Scale: `10`
- Headless: `false`

To target a non-default socket, pass `--socket` to every command:

```bash
./pxcli start --socket /tmp/alt.sock
./pxcli set_pixel 0 0 red --socket /tmp/alt.sock
./pxcli stop --socket /tmp/alt.sock
```

## Headless vs windowed

- Windowed mode is the default when built with `-tags=ebiten`.
- Headless mode is opt-in; pass `--headless` for CI or a headless container.
- If the binary is built without the `ebiten` tag, starting without `--headless` returns `err renderer_unavailable ...`.
- The headless container/CI environment cannot open a window; build and run windowed mode locally.

## Development

Typical commands:

```bash
go test ./...
go build ./cmd/pxcli
go build -tags=ebiten ./cmd/pxcli
```

If you are developing in a headless container, use `--headless` when running the daemon.

## Export path behavior

`pxcli export <filename.png>` resolves the filename to an absolute path before sending the request.
That means relative paths are interpreted from the CLI's current working directory,
not from where the daemon is running.

## Troubleshooting stale PID/socket files

If the daemon crashes or is killed, `/tmp/pxcli.pid` or `/tmp/pxcli.sock` may remain as stale pid/socket files:

- Re-run `pxcli start` (or `pxcli daemon`) to automatically remove stale pid/socket files.
- If you see `err daemon_already_running`, a live daemon is likely running on that socket.
  Use `pxcli stop --socket <path>` to shut it down, or point your CLI to the correct socket.

Manual deletion should not be necessary for normal recovery.
