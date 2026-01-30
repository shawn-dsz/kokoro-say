# kokoro-say

A fast CLI tool to convert text to speech using a local [Kokoro](https://github.com/remsky/Kokoro-FastAPI) TTS server.

## Features

- Near-instant startup (compiled Go binary)
- Pipe text or pass as argument
- Multiple voice options
- Adjustable speech speed
- Save to file or play directly
- **Web interface** - browser-based UI
- Zero dependencies beyond Go stdlib

## Requirements

- A running [Kokoro TTS server](https://github.com/remsky/Kokoro-FastAPI) (default: `http://localhost:8880`)
- One of these audio players (for playback):
  - `ffplay` (recommended, from FFmpeg)
  - `mpv`
  - `afplay` (macOS built-in)
  - `paplay` / `aplay` (Linux)

## Installation

### From source

```bash
go install github.com/shawn-dsz/kokoro-say@latest
```

Or clone and build:

```bash
git clone https://github.com/shawn-dsz/kokoro-say.git
cd kokoro-say
go build -o kokoro-say .
sudo mv kokoro-say /usr/local/bin/
```

### Pre-built binaries

Download from the [releases page](https://github.com/shawn-dsz/kokoro-say/releases).

## Usage

```bash
# Basic usage
kokoro-say "Hello, world!"

# Pipe text from another command
echo "Hello from a pipe" | kokoro-say
cat article.txt | kokoro-say

# Choose a voice
kokoro-say -v bf_emma "Hello with a British accent"

# Adjust speed (0.5 = half speed, 2.0 = double speed)
kokoro-say -s 1.5 "Speaking faster"

# Save to file instead of playing
kokoro-say -o greeting.mp3 "Hello, world!"

# List available voices
kokoro-say --voices

# Show help
kokoro-say --help

# Start web interface
kokoro-say web
```

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `-v` | Voice to use | `af_heart` |
| `-s` | Speech speed (0.5-2.0) | `1.0` |
| `-o` | Output file (skip playback) | - |
| `--voices` | List available voices | - |
| `--help` | Show help | - |

## Web Interface

Start a browser-based UI:

```bash
kokoro-say web
```

This opens a simple web page where you can paste text, select a voice, and click play.

| Flag | Description | Default |
|------|-------------|---------|
| `--port` | Port to listen on | `3456` |
| `--no-browser` | Don't auto-open browser | - |

**Keyboard shortcut:** `Cmd+Enter` (or `Ctrl+Enter`) to play.

## Available Voices

| Voice ID | Description |
|----------|-------------|
| `af_heart` | American female, warm (default) |
| `af_bella` | American female |
| `af_nicole` | American female |
| `af_sarah` | American female |
| `af_sky` | American female |
| `am_adam` | American male |
| `am_michael` | American male |
| `bf_emma` | British female |
| `bf_isabella` | British female |
| `bm_george` | British male |
| `bm_lewis` | British male |

## Configuration

Set `KOKORO_URL` environment variable to override the default server URL:

```bash
export KOKORO_URL="http://192.168.1.100:8880"
```

Default: `http://localhost:8880`

## How It Works

1. Accepts text from command-line arguments or stdin
2. Sends a POST request to the Kokoro TTS API (`/v1/audio/speech`)
3. Streams the audio response to your system's audio player (or saves to file)

The Kokoro server exposes an OpenAI-compatible TTS API, making it easy to integrate with existing tools.

## Troubleshooting

**"Kokoro server not found"**
- Ensure Kokoro TTS is running at the configured URL
- Check with: `curl http://localhost:8880/v1/voices`

**No audio playback**
- Install `ffplay` (from FFmpeg): `brew install ffmpeg` or `apt install ffmpeg`
- Or install `mpv`: `brew install mpv` or `apt install mpv`

**Slow startup**
- This shouldn't happen with the Go binary. If you're running from source with `go run`, use `go build` instead.

## License

MIT
