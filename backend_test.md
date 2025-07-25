# Backend Testing Roadmap (Intern Edition)

## 1. Objective

Deliver a **robust, high-coverage automated test suite** for the `backend/` Go codebase.  
Target ≥ **70 % statement coverage** (excluding generated code) while providing confidence in core business logic, handlers, and service layers.

---

## 2. Context & Documentation

*The intern must become self-sufficient; leverage Context7 for authoritative docs and snippets.*

| Dependency | Context7 Library ID | Suggested Topic |
|------------|--------------------|-----------------|
| Gin web framework | `/gin-gonic/gin` | Routing, testing with `httptest` |
| MongoDB Go Driver | `/mongodb/docs` | CRUD, aggregation mock patterns |
| Qdrant Go client | `/qdrant/qdrant-go-client` | Upsert, search APIs |
| Testify | `/stretchr/testify` | `assert`, `require`, mock package |
| Go stdlib | `/golang/go` | `testing`, `httptest`, `net/http/httptest` |

**Tip**: Use the Context7 search bar → _“/gin-gonic/gin testing”_ to fetch ready-made code examples.

---

## 3. High-Level Milestones

1. **Environment & Tooling** – prepare local dev scripts, CI hooks (Week 1)  
2. **Core Package Coverage** – handlers, services, models (Weeks 1-3)  
3. **Database & External Integrations** – in-memory fakes, error paths (Weeks 4-5)  
4. **End-to-End Smoke Tests** – minimal containerised Mongo/Qdrant (Week 6)  
5. **Coverage Optimisation & Refactor Tickets** – reach ≥ 70 % (Week 7)  
6. **Documentation & Handoff** – README updates, coverage badge (Week 8)

---

## 4. Tooling Setup

| Tool | Why / Notes |
|------|-------------|
| **Go 1.22+** | Match repo’s `go.mod` version |
| **Testify** | Assertions & mocking (`go get github.com/stretchr/testify`) |
| **GoMock or Hand-rolled fakes** | Optional for complex interfaces |
| **go-cover** & **go-tool cover** | HTML coverage reports |
| **GolangCI-Lint** | Ensure new tests pass linting |
| **Taskfile / Makefile** | Add `make test`, `make coverage` targets |
| **GitHub Actions** | CI gate on coverage % |

*Deliverable*: `scripts/test.sh` that runs unit tests, generates `coverage.html`, and fails if < 70 %.

---

## 5. Mocking Strategy

1. **Pure Interfaces** – wrap external deps:  
   ```go
   type MongoCollection interface {
       Aggregate(ctx context.Context, pipeline any, opts ...*options.AggregateOptions) (Cursor, error)
       InsertOne(ctx context.Context, doc any) (*mongo.InsertOneResult, error)
       // ...
   }
   ```
2. **Hand-written fakes** for simple behaviour (e.g. return canned results, record calls).
3. **Table-Driven Tests** – enumerate inputs/expected outputs.
4. **Golden Files** – for large JSON payloads (e.g. sample GitHub events).
5. **httptest.Server** – spin up Gin router with mocked services; assert JSON/HTTP codes.

---

## 6. Scope by Package

### 6.1 `internal/handlers` (≈2 100 LOC)

| Endpoint | Key Scenarios | Notes |
|----------|---------------|-------|
| `POST /auth/login` | correct creds, bad creds, disabled user | mock `AuthService` |
| `GET /search` | query param combos, repo filter, pagination | mock `SearchService` |
| `POST /chat` | prompt validation, streaming | mock `ChatRagService` |
| `POST /repos/:id/files` | large upload, invalid syntax | mock `RepositoryService` |

*Tasks*
1. Build `mockService.go` implementing required service interfaces.
2. Use `httptest.NewRecorder` to drive requests.
3. Assert status codes, JSON schema (use `assert.JSONEq`).

### 6.2 `internal/services` (≈4 280 LOC)

Breakdown & test matrix:

| File | Function | Happy Path | Error Cases | Edge | Needs Mock |
|------|----------|-----------|-------------|------|------------|
| `search.go` | `SearchCodeChunks` | ✓ | bad query, db error | offset>docs | mongodb, qdrant |
| `embedding_pipeline.go` | `ProcessRepository` | ✓ | provider fail | large repo | embedding provider |
| `github.go` | `SyncRepo` | ✓ | rate limit | private repo | GitHub API |
| `user.go` | `CreateUser`, `VerifyPassword` | ✓ | dup email | unicode pw | bcrypt |

*Approach*
- Provide tiny fake implementations of Mongo & Qdrant that store docs in memory.  
- Use channels or context deadlines to simulate latency/timeouts.

### 6.3 `internal/database` (662 LOC)

Integration-style unit tests with **mongo-memory-server** or **Testcontainers**.  
Focus: connection string parsing, retry logic, `UpsertPoints` query building.

### 6.4 `internal/server` (227 LOC)

- Start Gin engine with all middleware & routes.
- Hit `/health` and check 200 & JSON body.

### 6.5 `cmd/*`

- Replace `os.Args`, call `main()`, capture stdout/stderr.
- Verify flags parsing, error exits.

### 6.6 `internal/models` (+ remaining)

- Table-driven tests for helper fns: `NormalizeLanguage`, CRUD methods.

---

## 7. Exclusions (Do NOT spend time)

| Path | Rationale |
|------|-----------|
| `internal/generated/*` | Auto-generated OpenAPI stubs; exercise indirectly via handlers |
| `scripts/`, `Dockerfile*` | Execution tested in CI build, not unit tests |

---

## 8. Continuous Integration (CI)

*Add a new workflow `.github/workflows/test.yml`*
```yaml
ame: backend-test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 1.22
      - run: make test
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: backend/coverage.out
```

---

## 9. Milestone Checklist

- [ ] `make test` command returns non-zero on failures  
- [ ] Coverage ≥ 40 % by end of Week 3  
- [ ] Handlers & services mock coverage ≥ 70 %  
- [ ] CI green on pull requests  
- [ ] README updated with badges & instructions  

---

## 10. Tips & Pitfalls

1. **Avoid hitting real external services** – keep tests < 200 ms.  
2. **Keep mocks small** – prefer in-line fakes over heavy frameworks.  
3. **Refactor fearlessly** – if code resists testing, wrap with an interface.  
4. **Read Context7 first** – saves hours of trial-and-error.

---

Happy testing! Ping the mentor if you’re blocked for > 30 min. 