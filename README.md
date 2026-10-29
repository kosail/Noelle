# Noelle

Noelle is a small Slack bot written in Go that answers questions about local files and documentation using an AI backend.

Instead of building session management, context loading, and tool execution from scratch, Noelle talks directly to a local [OpenCode](https://opencode.ai) server. The idea is to keep things simple, practical, and easy to maintain for a small team.

## Why this exists

Our team has about 10 people, and at most 5 or 6 of us will probably be asking questions at the same time. We wanted a straightforward way to query our local project files directly from Slack without having to hunt down files or re-read tons of documentation manually.

Using Go and OpenCode makes this clean:
* OpenCode server handles the heavy lifting (sessions, chat history, file indexing, and tools).
* Go gives us a light, reliable single binary that connects Slack to OpenCode via Socket Mode.
* We do not need multi-platform support or swappable AI backends. Slack plus OpenCode does the job.

## Interaction Modes

Noelle supports three different modes depending on who is asking and what kind of context they need:

1. **Normal mode (`@Noelle`)**  
   Tag the bot directly in a channel or thread for general questions about project files and documentation.

2. **Artist mode (`/artist <question>`)**  
   Tailored for our Unity frontend artists and developers who focus on gameplay mechanics and visuals. It skips heavy architecture jargon and explains things in terms of Unity components, mechanics, and straightforward scripts.

3. **Developer mode (`/dev <question>`)**  
   Geared towards core software engineers. Answers dive into low-level code, engine internals, systems design, and infrastructure.

## Project Status

This project is in its early stages. Only the basic Slack connection (via Socket Mode) and OpenCode client setup are in place.

## Getting Started

### Prerequisites

* Go 1.22 or newer
* A running OpenCode server instance with access to your local files
* A Slack App with Socket Mode enabled

### Slack App Setup

You will need two tokens from the Slack developer dashboard:
* **Bot User OAuth Token** (`xoxb-...`): Needs permissions to read mentions and send messages (such as `app_mentions:read`, `chat:write`, `commands`).
* **App-Level Token** (`xapp-...`): With the `connections:write` scope for Socket Mode.

### Configuration

Copy the example environment file and fill in your details:

```bash
cp .env.example .env
```

Your `.env` should look like this:

```env
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_APP_TOKEN=xapp-your-app-token

OPENCODE_SERVER_IP=127.0.0.1
OPENCODE_SERVER_PORT=4096
OPENCODE_SERVER_PASSWORD=your-opencode-password
```

### Running

To run locally in development:

```bash
go run ./src
```

Or build a binary:

```bash
go build -o noelle ./src
./noelle
```

## License

MIT. Free to use, modify, and distribute.

---
2026, kosail. With love, from Honduras.
