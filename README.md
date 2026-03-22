# BurstUI

A terminal UI for Gobuster built with Bubble Tea and Lip Gloss.

BurstUI provides a small TUI around common Gobuster workflows so you can switch modes, set options, run scans, and review live output without typing the full command each time.

Current release version: `v0.1.0`

## Features

- Supports Gobuster `dir`, `vhost`, and `dns` modes
- Startup detection of Gobuster path and version
- Live scan output with scrollable result pane
- Thread count input
- Mode-aware status code filters
- Wordlist file picker
- Optional custom DNS resolver for `dns` mode
- Output log export after a scan completes

## Requirements

- Docker

OR

- Go `1.26.1` or newer
- Gobuster `3.8.2` or newer
- A working `gobuster` binary in your `PATH`

BurstUI checks Gobuster on startup and prints both the resolved path and detected version in the output panel.

## Use with Docker

### Pull the latest image

```bash
docker pull ghcr.io/ctzisme/burstui:latest
```

### Run the TUI:

Run BurstUI with host wordlists mounted read-only and the current directory mounted to `/output` for saving scan logs.

```bash
docker run --rm -it \
  -e TERM=xterm-256color \
  -v /usr/share/wordlists:/usr/share/wordlists:ro \
  -v "$(pwd):/output" \
  ghcr.io/ctzisme/burstui:latest
```
Make sure to save logs only to the mounted directory (for example, `/output/123.log` when using the command above).

## Manual Installation

### Install BurstUI

Build from source:

```bash
go build -o burstui .
```

Run directly without creating a binary first:

```bash
go run .
```

### Install Gobuster with Go

Install the latest Gobuster release:

```bash
go install github.com/OJ/gobuster/v3@latest
```

Install the recommended Gobuster version `v3.8.2`:

```bash
go install github.com/OJ/gobuster/v3@v3.8.2
```

If `GOBIN` is not set, `go install` places binaries in:

```text
$(go env GOPATH)/bin
```

On your machine that path is:

```text
/home/USER/go/bin
```

### Add Gobuster to PATH

Use it immediately in the current shell:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

On your machine, that is equivalent to:

```bash
export PATH="/home/USER/go/bin:$PATH"
```

Verify that the Go-installed Gobuster is the one being used:

```bash
which gobuster
gobuster --version
```

### Make the PATH change persistent

For Bash, add this line to `~/.bashrc`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

To also cover login shells, add the same line to `~/.profile`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Then reload your shell:

```bash
source ~/.bashrc
hash -r
```

## Usage

Start the app:

```bash
./burstui
```

When BurstUI starts, it will show:

- Gobuster path from `which gobuster`
- Gobuster version from `gobuster --version`

## Modes

### `dir`

Fields:

- Target URL
- Filter Status Codes
- Threads
- Wordlist Path

Generated command shape:

```bash
gobuster dir -u https://example.com -w wordlist.txt -s 200,301,302 -t 10 --status-codes-blacklist ""
```

Notes:

- You can set multiple `Filter Status Codes`, separated by commas (e.g. `200,304,403`).
- If left empty, BurstUI uses `200`

### `vhost`

Fields:

- Target URL
- Exclude Status Codes
- Threads
- Wordlist Path

Generated command shape:

```bash
gobuster vhost -u https://example.com -w wordlist.txt -t 10 -xs 400 --append-domain
```

Notes:

- You can set multiple `Exclude Status Codes`, separated by commas (e.g. `400,500`).
- If left empty, BurstUI uses `400`

### `dns`

Fields:

- Domain
- Custom DNS Server (Optional)
- Threads
- Wordlist Path

Generated command shape:

```bash
gobuster dns -do example.com -w wordlist.txt -t 10
```

With a custom resolver (optional):

```bash
gobuster dns -do example.com -w wordlist.txt -t 10 --resolver 8.8.8.8:53
```

Notes:

- You cannot filter status codes in `dns` mode
- `Custom DNS Server` is optional

## Controls

- `↑ / ↓`: move between fields
- `← / →` on `Mode`: switch mode
- `→`: fill a field with its placeholder when empty
- `Tab` / `Shift+Tab`: switch between form and result pane
- `Enter` on `Browse Wordlist`: open the wordlist picker
- `Enter` on `Start Scan`: run Gobuster
- `Enter` on `Output Log File`: save logs after a scan completes
- `ctrl+c`: quit

## Output Logs

After a scan finishes, you can save the collected output to a file using the `Output Log File` field.

Default output path:

```text
./burstui-output.log
```

## Project Structure

- `main.go`: program entrypoint
- `model.go`: model, initialization, Gobuster startup detection
- `update.go`: key handling, focus logic, update loop
- `scan.go`: Gobuster command construction and output streaming
- `view.go`: TUI rendering
- `styles.go`: Lip Gloss styles and log coloring helpers
