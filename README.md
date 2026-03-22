# BurstUI

A Go terminal UI for Gobuster built with Bubble Tea and Lip Gloss.

BurstUI provides a small TUI around common Gobuster workflows so you can switch modes, set options, run scans, and review live output without typing the full command each time.

[![License](https://img.shields.io/github/license/ctzisme/burstui)](https://github.com/ctzisme/burstui/blob/main/LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/ctzisme/burstui)](https://github.com/ctzisme/burstui/releases)
[![Go Version](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go)](https://go.dev/)
[![Gobuster](https://img.shields.io/badge/Gobuster-3.8.2-orange)](https://github.com/OJ/gobuster)
[![GHCR](https://img.shields.io/badge/GHCR-ghcr.io%2Fctzisme%2Fburstui-blue)](https://github.com/ctzisme/burstui/pkgs/container/burstui)
[![Stars](https://img.shields.io/github/stars/ctzisme/burstui?style=social)](https://github.com/ctzisme/burstui)

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

*OR*

- Go `1.26.1` or newer
- Gobuster `3.8.2` or newer
- A working `gobuster` binary in your `PATH`

## 🐳 Use with Docker

### Pull the latest image

```bash
docker pull ghcr.io/ctzisme/burstui:latest
```

### Run a container

Run a BurstUI container with host wordlists mounted read-only and the current directory mounted to `/output` for saving scan logs.

```bash
docker run --rm -it \
  -e TERM=xterm-256color \
  -v /usr/share/wordlists:/usr/share/wordlists:ro \
  -v "$(pwd):/output" \
  ghcr.io/ctzisme/burstui:latest
```
*Make sure to save logs only to the mounted directory (for example, `/output/123.log` when using the command above).*

## 🔨 Manual Installation

### Step 1. Install and build BurstUI

#### Option A: Using Go Install
```bash
go install github.com/ctzisme/burstui@latest
```

*OR*

#### Option B: From Source Code
```bash
git clone https://github.com/ctzisme/burstui
cd burstui
go mod tidy
go build .
```
*OR*

#### Option C: Using Binary Releases

Download binary releases from the [releases page](https://github.com/ctzisme/burstui/releases).

### Step 2. Install Gobuster with Go

Install the recommended Gobuster version `v3.8.2`:

```bash
go install github.com/OJ/gobuster/v3@v3.8.2
```

*OR*

Install the latest Gobuster release:

```bash
go install github.com/OJ/gobuster/v3@latest
```

### Step 3: Add the latest Gobuster to PATH

Use it immediately in the current shell:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Verify that the Go-installed Gobuster is the one being used:

```bash
which gobuster
gobuster --version
```

### Step 4: Make the PATH change persistent

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

## Controls

- `↑ / ↓`: move between fields
- `← / →` on `Mode`: switch mode
- `→`: fill a field with its placeholder when empty
- `Tab` / `Shift+Tab`: switch between form and result pane
- `Enter` on `Browse Wordlist`: open the wordlist picker
- `Enter` on `Start Scan`: run Gobuster
- `Enter` on `Output Log File`: save logs after a scan completes
- `ctrl+c`: quit

## Modes

### `dir` (Directory Mode)

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

### `vhost` (Virtual Host Mode)

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

### `dns` (DNS Mode)

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

## Output Logs

After a scan finishes, you can save the collected output to a file using the `Output Log File` field.

Default output path:

```text
./burstui-output.log
```
*Make sure to save logs only to the mounted directory if using Docker.*

## Project Structure

- `main.go`: program entrypoint
- `model.go`: model, initialization, Gobuster startup detection
- `update.go`: key handling, focus logic, update loop
- `scan.go`: Gobuster command construction and output streaming
- `view.go`: TUI rendering
- `styles.go`: Lip Gloss styles and log coloring helpers
- `version.go`: version metadata variables for release builds
