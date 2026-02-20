# AgentZero

Terminal chat client built with Go, Bubble Tea, and LangChainGo/OpenAI.

## Features

- Interactive terminal UI (Bubble Tea)
- Conversation history within a session
- Simple slash commands (`/clear`, `/quit`, `/exit`)
- Optional custom OpenAI base URL and model via environment variables

## Requirements

- Go 1.25+
- OpenAI-compatible API key available to your environment

## Setup

1. Clone the repository.
2. Create a `.env` file (or export env vars directly).
3. Set the required credentials/config.
4. Run the app.

Example `.env`:

```env
OPENAI_API_KEY=your_api_key_here
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o-mini
```

`OPENAI_BASE_URL` and `OPENAI_MODEL` are optional; defaults are used when omitted.

## Run

```bash
go run .
```

Or build first:

```bash
go build -o agentzero .
./agentzero
```

## In-App Controls

- `Enter`: send message
- `/clear`: clear chat history for current session
- `/quit` or `/exit`: quit
- `Esc` / `Ctrl+C`: quit

## Notes

- `.env` is loaded automatically if present.
- If no `.env` file exists, the app still runs and uses exported environment variables.
