# kokoro-say

A simple CLI tool to convert text to speech using a local Kokoro TTS server.

## Overview

`kokoro-say` is a command-line utility that sends text to a locally running Kokoro TTS server and plays the audio. It's designed to be simple, fast, and easy to use.

## Features

- Pipe text or pass as argument
- Choose from multiple voices
- Adjust speech speed
- Save to file instead of playing
- List available voices
- Configurable server URL via environment variable

## Installation

```bash
npm install -g kokoro-say
# or
brew install kokoro-say  # future goal
```

## Usage

```bash
# Basic usage
kokoro-say "Hello, world!"

# Pipe text
echo "Hello from a pipe" | kokoro-say

# Choose a voice
kokoro-say -v bf_emma "Hello with a British accent"

# Adjust speed (0.5 = half speed, 2.0 = double speed)
kokoro-say -s 1.2 "Speaking a bit faster"

# Save to file instead of playing
kokoro-say -o greeting.mp3 "Hello, world!"

# List available voices
kokoro-say --voices

# Show help
kokoro-say --help
```

## Configuration

Set `KOKORO_URL` environment variable to override the default server URL:

```bash
export KOKORO_URL="http://localhost:8880"
```

Default: `http://localhost:8880`

## Technical Plan

### Language Choice

**Node.js/TypeScript** - Good cross-platform support, easy npm distribution, handles HTTP and audio playback well.

Alternative: **Go** - Single binary, no runtime dependencies. Consider if distribution simplicity is priority.

### Project Structure

```
kokoro-say/
├── src/
│   └── index.ts          # Main CLI entry point
├── package.json
├── tsconfig.json
├── README.md
├── LICENSE               # MIT
└── .gitignore
```

### Dependencies

- `commander` - CLI argument parsing
- `node-fetch` or built-in fetch (Node 18+) - HTTP requests
- Audio playback: spawn `afplay` (macOS), `aplay` (Linux), or `powershell` (Windows)

### Implementation Steps

1. **Setup project**
   - Initialize package.json with bin entry
   - Configure TypeScript
   - Add .gitignore

2. **Implement core functionality**
   - Parse CLI arguments with commander
   - Read input from args or stdin
   - POST to Kokoro API endpoint `/v1/audio/speech`
   - Handle response (audio stream)

3. **Audio playback**
   - Detect platform (darwin/linux/win32)
   - Spawn appropriate player:
     - macOS: `afplay -`
     - Linux: `aplay -` or `paplay`
     - Windows: Save temp file, play with PowerShell

4. **Additional features**
   - `--voices` flag: GET `/v1/voices` or hardcode known voices
   - `--output` flag: Write to file instead of playing
   - `--speed` flag: Pass speed parameter to API

5. **Polish**
   - Error handling (server not running, invalid voice, etc.)
   - Helpful error messages
   - README with examples

### API Reference

Kokoro exposes an OpenAI-compatible TTS API:

```
POST /v1/audio/speech
Content-Type: application/json

{
  "input": "Text to speak",
  "voice": "af_heart",
  "speed": 1.0,
  "response_format": "mp3"
}

Response: audio/mpeg stream
```

### Voice Options

| Voice ID | Description |
|----------|-------------|
| af_heart | Female, warm (default) |
| af_bella | Female, American |
| af_nicole | Female, American |
| af_sarah | Female, American |
| af_sky | Female, American |
| am_adam | Male, American |
| am_michael | Male, American |
| bf_emma | Female, British |
| bf_isabella | Female, British |
| bm_george | Male, British |
| bm_lewis | Male, British |

### Error Handling

- Server not running → "Kokoro server not found at {url}. Is it running?"
- Invalid voice → "Unknown voice '{voice}'. Use --voices to see available options."
- Empty input → "No text provided. Pass text as argument or pipe to stdin."

## Future Ideas

- `--interactive` mode for REPL-style input
- `--clipboard` to read from clipboard
- Config file (~/.kokoro-say.json) for default voice/speed
- Shell completions
