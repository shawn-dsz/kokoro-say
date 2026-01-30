# Architecture: kokoro-say

> This document is automatically updated after each checkpoint.
> Last updated: 2026-01-30 10:00

## Overview
A command-line utility for text-to-speech using a local Kokoro TTS server. Core flow: accept text input (argument or stdin) → POST to Kokoro API → stream audio → play via system player.

**Primary use case:** Quick text-to-speech for personal productivity.

## Components
| Component | Responsibility |
|-----------|---------------|
| CLI Parser | Parse flags (-v, -s, -o, --voices), read positional args |
| Input Handler | Read text from args or stdin |
| TTS Client | POST to Kokoro API, stream response |
| Audio Player | Platform-specific playback (afplay/aplay) |

## Data Flow
```
[CLI Args or Stdin]
        ↓
    [main.go]
        ↓
[POST /v1/audio/speech] → JSON: {input, voice, speed}
        ↓
  [audio/mpeg stream]
        ↓
  -o flag? ─┬─ yes → [write to file]
            └─ no  → [pipe to afplay/aplay]
```

## Technology Decisions
| Decision | Choice | Rationale | Alternatives Considered |
|----------|--------|-----------|------------------------|
| Language | Go | Near-instant startup (<10ms), single binary | Node.js (50-100ms startup) |
| CLI parsing | stdlib `flag` | Zero deps, fast | commander, cobra |
| HTTP client | stdlib `net/http` | Built-in, sufficient | resty |
| Audio playback | System player (afplay/aplay) | Native, no deps | portaudio bindings |

## API Contracts
*Pending: Will be populated during Build phase*

## File Structure
*Pending: Will be populated during Build phase*

## Security Considerations
*Pending: Will be populated after Review phase*

## Performance Considerations
*Pending: Will be populated after Review phase*

## Change Log
| Phase | Checkpoint | Change | Timestamp |
|-------|------------|--------|-----------|
| Setup | Initial | Created architecture document | 2026-01-30 10:00 |
