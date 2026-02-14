# pxcli

pxcli is a cli tool meant for AI agents (claude code, opencode, codex) to draw pixel art.

It was created using a ralph loop see `ralph.sh` and an example agent is given in `.opencode/agents/artists.md`

## Install (Homebrew)

macOS and Linux (Linuxbrew):

```bash
brew install vossenwout/pixel-art-cli/pxcli
```

Upgrade:

```bash
brew upgrade pxcli
```

## Build from source

Windowed (GUI):
(don't have windows PC so didn't tests this if you have fixes you can always make PR)

```bash
go build -tags=ebiten ./cmd/pxcli
```

Headless (no GUI deps):

```bash
go build ./cmd/pxcli
```

## Quick start

```bash
pxcli start
pxcli set_pixel 1 1 #ff0000
pxcli export out.png
pxcli stop
```

## CLI API

Lifecycle:

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

!!! For zsh shells you have to put colors between "" parenthesis. !!!

## Headless vs windowed

- Windowed mode is the default when built with `-tags=ebiten`.
- Headless mode is opt-in; pass `--headless` for CI or a headless container.
- If the binary is built without the `ebiten` tag, starting without `--headless` returns `err renderer_unavailable ...`.
- The headless container/CI environment cannot open a window; build and run windowed mode locally.

## Linux GUI requirements

For the GUI on Linux, you need X11/OpenGL runtime libraries. Examples:

- Debian/Ubuntu: `sudo apt-get install -y libx11-6 libxext6 libxrandr2 libxinerama1 libxcursor1 libxi6 libgl1`
- Fedora: `sudo dnf install libX11 libXext libXrandr libXinerama libXcursor libXi mesa-libGL`

If you are in a headless container, use `--headless`.

## Development

Typical commands:

```bash
go test ./...
go build ./cmd/pxcli
go build -tags=ebiten ./cmd/pxcli
```

If you are developing in a headless container, use `--headless` when running the daemon.
