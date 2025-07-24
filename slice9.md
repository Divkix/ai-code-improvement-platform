# Slice 9 – AI Chat Interface

> Goal: deliver a fully-functional conversational interface that lets users ask natural-language questions about their code repositories, powered by a Retrieval-Augmented-Generation (RAG) pipeline using Claude 4 Sonnet.
>
> Time-box: **3 days (24 work-hours)**
>
> Success criteria
> 1. Users can open the Chat page, pick a repository and send messages.
> 2. Backend returns context-aware answers that quote specific code lines.
> 3. Average first-token latency ≤ 5 s for repo ≤ 3 k chunks.
> 4. Chat session history is saved and reloads correctly.
> 5. Mobile layout & basic accessibility (a11y) pass.

---

## 0  Prerequisites (verify before starting)
- Voyage embedding + vector search working (Slice 8 complete).
- `ANTHROPIC_API_KEY` available in `.env` & Docker Compose.
- Frontend route `/chat` scaffold exists.

---

## 1  Database & Models (3 h)
1.1 Create **`chat_sessions`** collection in MongoDB.
```go
// internal/models/chat_session.go
// ... fields: _id, userId, repositoryId, title, messages[], createdAt, updatedAt
```
1.2 Define **`ChatMessage`** sub-document schema: role, content, retrievedChunks[], tokensUsed, timestamp.
1.3 Add helper methods: `NewChatSession`, `AppendMessage`, `ListSessionsByUser`.
1.4 Write unit tests for model helpers (Go test).

---

## 2  LLM Provider – OpenAI-compatible API (2 h)
2.1 **Dependency**: `github.com/sashabaranov/go-openai` (MIT) — official Go SDK for any OpenAI-schema endpoint.  
   • Works with OpenAI, Azure OpenAI, Groq, Together.ai, Fireworks, OpenRouter, etc. via configurable `BaseURL`.
2.2 Add `internal/services/llm.go` abstraction:
```go
// ChatStream returns a channel of streamed chat chunks
func (c *Client) ChatStream(ctx context.Context, messages []openai.ChatCompletionMessage, opts ChatOpts) (<-chan openai.ChatCompletionStreamResponse, error)
```
   – Wraps go-openai client with retry + timeout.
2.3 Environment variables (document in README):
```bash
LLM_BASE_URL=https://api.openai.com/v1      # Default; override for Groq, Together, etc.
LLM_MODEL=gpt-4o-mini                       # Or claude-3-sonnet via Groq
LLM_API_KEY=sk-...
LLM_REQUEST_TIMEOUT=30s
```
2.4 Update Docker Compose to inject these vars into backend service.
2.5 Remove old Anthropic-specific references (ANTHROPIC_API_KEY, CLAUDE_MODEL).

---

## 3  Prompt Templates (1 h)
3.1 Create `internal/prompts/chat_prompt.go` with a Go template:
```
You are an AI assistant helping developers understand code. Use ONLY the provided code snippets.

{{range .Snippets}}
--- File: {{.FilePath}} ({{.StartLine}}-{{.EndLine}})
{{.Content}}
{{end}}

User question: {{.Question}}

Guidelines:
- Reference file paths & line numbers.
- Be concise.
- If unsure, say "I’m not certain".
```
3.2 Unit test rendering with fake data.

---

## 4  Backend API (6 h)
4.1 OpenAPI additions (`backend/api/openapi.yaml`):
```
POST /api/chat/sessions             # create session (optional repoId)
GET  /api/chat/sessions             # list sessions
GET  /api/chat/sessions/{id}        # get session details
POST /api/chat/sessions/{id}/message  # send user message, stream assistant response
DELETE /api/chat/sessions/{id}      # delete session
```
4.2 Regenerate server stubs (`make generate`).
4.3 Handlers (`internal/handlers/chat.go`):
- `CreateSessionHandler`
- `ListSessionsHandler`
- `GetSessionHandler`
- `PostMessageHandler` (core RAG flow)
4.4 `services/chat_rag.go`: orchestrates
```
1) embed & vector search (8 nearest chunks)
2) build prompt template
3) call Anthropic client (stream)
4) persist assistant & user messages
```
4.5 Streaming via `gin.Context.Stream` + SSE.
4.6 Middleware: ensure user owns repository.
4.7 Integration tests with in-memory Mongo/Qdrant mocks.

---

## 5  Frontend – SvelteKit (6 h)
5.1 Store: `src/stores/chat.ts` (selectedSession, messages, isLoading).
5.2 API utility: `lib/api/chat-client.ts` wrapping endpoints & EventSource for streaming.
5.3 Update `/routes/chat/+page.svelte`:
- Session list sidebar (mobile drawer).
- Message list with role-based styling.
- Code blocks rendered with Prism.
- Typing indicator while streaming.
- Suggested question chips when no messages.
5.4 `Message.svelte` component – supports copy-code button.
5.5 `RepositorySelector` integration at top bar.
5.6 Basic responsive CSS (Tailwind classes).

---

## 6  Metrics & Logging (1 h)
6.1 Add Prometheus counters: total chat requests, Claude tokens, latency histogram.
6.2 Structured logs: sessionId, vectorHits, latencyMs.

---

## 7  E2E & QA (3 h)
7.1 Playwright test: import sample repo → open chat → ask question → expect answer contains code path.
7.2 Vitest unit tests for Svelte store.
7.3 Load test: 20 parallel chat requests, ensure P95 latency < 5 s.

---

## 8  Docs & Handoff (1 h)
8.1 Update `README.md` Chat section: env vars, run instructions.
8.2 Add OpenAPI example request/response.
8.3 Create demo script snippet.

---

## Deliverables Checklist
- [ ] Mongo model & migrations
- [ ] Anthropic client with retries
- [ ] Prompt template & tests
- [ ] OpenAPI spec updated + regenerated code
- [ ] Chat handlers & service layer
- [ ] Streaming SSE responses
- [ ] Frontend chat UI completed
- [ ] E2E & unit tests green
- [ ] Metrics exposed at `/metrics`
- [ ] Docs updated

---

## Nice-to-Have (do only if time remains)
- Feedback thumbs-up/down stored on messages.
- Syntax-highlight specific language per snippet.
- AI-generated session titles.
- Conversation export to Markdown. 