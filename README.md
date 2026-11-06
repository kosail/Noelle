![noelle logo](.github/images/banner.webp)

Noelle is a small Slack bot written in Go that answers questions about local files and documentation using an AI backend.

Instead of building session management, context loading, and tool execution from scratch, Noelle talks directly to a local [OpenCode](https://opencode.ai) server. The idea is to keep things simple, practical, and easy to maintain for a small team.

> [!NOTE]
> This project was built entirely by me, except for the icon of this project, the `context/` system prompts and the indexes in `docs/`.
> These were built iteratively and heavily optimized following several RAG and agentic design recommendations.
>
> I will leave the links of sources used at the bottom of this file, so you can consult them too. 


## 🌹 Why this exists

In the company that I work for, our team has about 10 people and thus we usually have not a lot of time to keep checking on documentation. Though, as our processes are highly documented and the issue is just time, I decided to implement the missing node: a straightforward way to query our local project files directly from Slack without having to hunt down files docs manually.

Using Go and OpenCode makes this clean:
* OpenCode server handles the heavy lifting (sessions, chat history, file indexing, and tools).
* Go gives us a light, reliable single binary that connects Slack to OpenCode via Socket Mode, and which would be cheap to run in AWS Graviton.
* We do not need multi-platform support or swappable AI backends. Slack plus OpenCode does the job.

The name of this project came from one of my favourites characters of a videogame I like, which is "Noelle". She is a warrior, and a maid. I just felt it was kinda funny to call a project like it, and so that's it.

## 🪷 Interaction Modes

Noelle supports three different modes depending on who is asking and what kind of context they need:

1. **Normal mode (`@Noelle`)**  
   Tag the bot directly in a channel or thread for general questions about project files and documentation.

2. **Artist mode (`/artist <question>`)**  
   Tailored for our Unity frontend artists and developers who focus on gameplay mechanics and visuals. It skips heavy architecture jargon and explains things in terms of Unity components, mechanics, and straightforward scripts.

3. **Developer mode (`/dev <question>`)**  
   Geared towards core software engineers. Answers dive into low-level code, engine internals, systems design, and infrastructure.

> [!IMPORTANT]
> Noelle creates one OpenCode session per Slack thread, not per channel or per project. A top-level @Noelle message starts a new Slack thread and an OpenCode session tied to it. Replies in that thread reuse it. Different threads in the same channel get different sessions. 
>
> You gotta notice that `/artist` and `/dev` create a new announcement thread (this is how Slack works), so each command gets its own session.
>
> Idle sessions are cleaned up after one hour, but only cleaned up in Noelle's memory, and never from OpenCode. This automatic OpenCode cleanup may be relevant in future releases.

## 🌷 Getting Started

### Prerequisites

* Go 1.22 or newer
* A running OpenCode server instance with access to your local files
* A Slack App with Socket Mode enabled

### Slack App Setup

As there is a lot of documentation online, I will not cover how to configure a Slack bot. In summary, you will need two tokens from the Slack developer dashboard:
1. **Bot User OAuth Token** (`xoxb-...`): Needs permissions to read mentions and send messages (such as `app_mentions:read`, `chat:write`, `commands`).
2. **App-Level Token** (`xapp-...`): With the `connections:write` scope for Socket Mode.

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

# The model provider to use in OpenCode (e.g. "opencode-go")
# Optional. If not set, fallbacks to system default.
# If set, needs OPENCODE_MODEL to be present or else this is ignored
OPENCODE_PROVIDER=

# The model to use (e.g. deepseek-v4.1-flash)
# Optional. If not set, fallbacks to system default.
# If set, needs OPENCODE_PROVIDER to be present or else this is ignored
OPENCODE_MODEL=
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

## 🌺 Adding Project Documentation

Noelle does not build or update documentation indexes at runtime. The approach taken here is that **you own your docs, and Noelle cannot modify them in any sort of way**. So building or updating your docs or anything related to it is a work that must happen outside of this bot lifecycle.

I understand that many projects may have different documentation structures, and/or may be located in many places. Thus, the approach this bot takes is an index-based one that helps route the agent quickly and converge on results fast.

Noelle expects these indexes to be in the `docs/` folder in this way:
* `docs/<PROJECT_FOLDER>/index.md` an index for the project itself.
* `docs/index.md` for the global project list.

To add a project, you need an AI model with access to the project files/docs and the templates in `docs/`.

1. Give the AI the project documentation and `docs/_project/index.md.example`.
2. Ask it to do a deep analysis of the docs and build the index. It should find topics that appear in several files, pick one primary source for each topic, and point to the other files that support it.
3. Save the result as `docs/<PROJECT_FOLDER>/index.md`.
4. Give the AI the new project index and `docs/index.md.example`.
5. Ask it to update `docs/index.md` with the new project route, cross-project topics, relationships, and index health.
6. Check that the paths and section headings exist. Try a few real questions to confirm the routes.

For this analysis procedure filenames alone are not enough. The AI needs to read the docs, because the same topic can live in several files, and we are precisely trying to get into that situations so we can register those results.

On runtime, Noelle reads the global index first, routes to the specific project, and based in the project index, opens only the files needed to answer the question.


## 🌸 Resources used
- [Anthropic, Effective context engineering for AI agents](https://www.anthropic.com/research/effective-context-engineering-for-ai-agents)
- [Anthropic, Contextual retrieval](https://www.anthropic.com/research/contextual-retrieval)
- [OpenAI, Retrieval](https://developers.openai.com/api/docs/guides/retrieval)
- [Google Cloud, Filter with metadata search](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/rag-engine/use-metadata-search)
- [Microsoft, Develop an agentic RAG solution](https://learn.microsoft.com/en-us/azure/architecture/ai-ml/guide/rag/rag-agentic)
- [AWS, Agentic retrieval best practices](https://docs.aws.amazon.com/wellarchitected/latest/agentic-ai-lens/agentperf03-bp05.html)
- [Agentic Design, Hierarchical Index Retrieval](https://agentic-design.ai/patterns/knowledge-retrieval/hierarchical-index-retrieval)

### Fonts
- [Frauces, from Google Fonts](https://fonts.google.com/specimen/Fraunces). Used in the banner for this repository.

---

## 💐 Contributing

Contributions are welcome!
Feel free to fork the repository and submit pull requests.
If you have ideas, suggestions, or bug reports, open an issue on GitHub.


## 📜 License

[MIT License](LICENSE), which means you are free to use, modify, and distribute.

---
Noelle. 2026, kosail.
<br>
With love, from Honduras.

