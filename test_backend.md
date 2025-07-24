# Backend Unit-Testing Roadmap

> **Audience**: New intern tasked with improving and maintaining unit-tests for `ai-code-improvement-platform/backend`
>
> **Scope**: Pure unit-tests (no network, no real DB). Integration tests are mentioned only where helpful.

---

## 1. Why We Test

1. Catch regressions early and with confidence.
2. Document intended behaviour – future maintainers can read tests as examples.
3. Enable fearless refactors – a green test-suite is our safety-net.
4. Provide quality gates in CI/CD (pull-requests must keep coverage ≥ threshold).

## 2. Current State (May 2025)

| Package                               | Tests | Cov. | Notes |
| ------------------------------------- | :---: | :--: | ----- |
| `internal/models`                      | ✅    |14 %| good start, needs more edge-cases |
| `internal/prompts`                     | ✅    |87 %| almost complete |
| `internal/services`                    | ✅    |1 % | only `llm_test.go` exists |
| *all others* (`auth`, `handlers`, …)   | ❌    |0 % | no unit-tests yet |
| **Overall**                            |       |~8 %| far from target |

_Command used: `go test ./... -cover` (see script in §6)_

## 3. Target & Milestones

| Milestone | Coverage goal | ETA |
|-----------|---------------|-----|
| M1: Enable tooling / CI   | ≥ 10 % (baseline) | Week 1 |
| M2: Core business logic   | ≥ 35 % | Week 2 – 3 |
| M3: Handlers & Middleware | ≥ 50 % | Week 4 – 5 |
| M4: Persistence boundary  | ≥ 60 % | Week 6 – 8 |

> 60 % strikes a balance between effort & ROI. Higher is welcome, but diminishing returns past ~80 %.

## 4. Tooling & Conventions

### 4.1 Standard Library
* `testing` – always available; table-driven + sub-tests.
* `httptest` – HTTP handler testing.
* `embed` – supply fixture files inline.

### 4.2 External Helpers (add to `go.mod`)

```bash
# one-time setup
go get github.com/stretchr/testify@v1.8.4   # assertions
go get github.com/golang/mock/gomock@v1.6.0  # mocking framework
# code-generation helper (install globally)
go install github.com/golang/mock/mockgen@latest
```

Why these?
* **Testify** gives expressive `assert.Equal`, `require.NoError` etc. – improves readability.
* **GoMock / mockgen** auto-generates mocks from interfaces (e.g. `database.Client`, `services.EmbeddingProvider`).

> NOTE: no third-party libs are in use today, so remember to run `go mod tidy` after adding them.

### 4.3 Makefile targets (add if missing)

```makefile
TEST_FLAGS ?= -race -count=1 -v

test:
	go test ./... $(TEST_FLAGS)

test-cover:
	go test ./... $(TEST_FLAGS) -coverprofile=coverage.out
	go tool cover -func=coverage.out  # human summary
```

### 4.4 Folder & File Rules

* A test file lives next to source file, suffixed with `_test.go`.
* External-facing tests (black-box) may use `package foo_test` to avoid touching internals.
* Table-driven style for multiple cases.
* Use `t.Parallel()` inside non-mutating sub-tests.
* Keep mocks small – favour fakes via interfaces over real DB.


## 5. Package-by-Package Checklist

### 5.1 `internal/auth`
* **Targets**: `HashPassword`, `ComparePassword`, `GenerateJWT`, `ValidateToken`.
* **Approach**: Provide deterministic inputs; assert expected hash prefix, error cases for wrong password, expired token etc.

### 5.2 `internal/config`
* **Targets**: `LoadConfig`.
* **Approach**: Use `os.Setenv` to craft env-vars; test default fall-backs and invalid duration parsing.

### 5.3 `internal/middleware`
* **Targets**: `AuthMiddleware`.
* **Approach**: wrap a dummy handler, inject mock auth-service via interface, use `httptest` recorder.

### 5.4 `internal/handlers`
* **HTTP handlers**: `health.go`, `dashboard.go`, `search.go`, …
* **Approach**: For each handler:
  1. Spin up recorder/request.
  2. Inject mock services (e.g. `SearchService`) crafted with GoMock.
  3. Assert status code & JSON body.

### 5.5 `internal/services`
| Sub-package | Key functions | Notes |
|-------------|---------------|-------|
| `embedding.go`          | `CountTokens`, `BuildEmbeddingInput` | pure – easy unit test |
| `local_embedding.go`    | off-disk model; can be mocked |
| `github.go`             | wraps GitHub API – **mock** interface |
| `chat_rag.go`           | orchestrates RAG pipeline – split into smaller functions & mock providers |

Refactor tightly-coupled code to depend on small interfaces so that they are mockable.

### 5.6 `internal/database`
* **Unit tests** should mock Mongo/Qdrant clients – **do not** hit real DB.
* Provide interface wrappers (`type MongoClient interface { Database(name string) *mongo.Database }`) and generate mocks.


## 6. Scripts & Automation

`backend/scripts/coverage.sh`  (suggested)
```bash
#!/usr/bin/env bash
set -euo pipefail
cd $(dirname "$0")/..
PROFILE=coverage.out
go test ./... -coverprofile=$PROFILE -covermode=atomic
if [ -f "$PROFILE" ]; then
  go tool cover -func=$PROFILE | tail -n1
fi
```

Run with `make test-cover` locally; CI will execute the same script.

## 7. Continuous Integration (GitHub Actions)

```
name: backend-tests
on:
  pull_request:
    paths:
      - 'backend/**'
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: make test-cover
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: backend/coverage.out
```

CI fails if coverage drops below maintained threshold (set via Codecov YAML or `go tool cover -func` check).

## 8. Example Test Patterns

### 8.1 Table-Driven + Sub-tests
```go
func TestCountTokens(t *testing.T) {
  t.Parallel()
  cases := []struct{
    name string
    input string
    want int
  }{{"empty", "", 0}, {"simple", "hello", 1}, {"sentence", "hello world", 2}}

  for _, tc := range cases {
    tc := tc // capture
    t.Run(tc.name, func(t *testing.T) {
      t.Parallel()
      got := CountTokens(tc.input)
      assert.Equal(t, tc.want, got)
    })
  }
}
```

### 8.2 Mocking with GoMock
```go
ctrl := gomock.NewController(t)
mockEmb := mocks.NewMockEmbeddingProvider(ctrl)
mockEmb.EXPECT().Embed(gomock.Any(), "hello").Return(vec, nil)
svc := services.NewChatRAG(mockEmb, /* other deps */)
```

### 8.3 Handler Testing
```go
req := httptest.NewRequest(http.MethodGet, "/health", nil)
rec := httptest.NewRecorder()
handlers.Health(rec, req)
require.Equal(t, http.StatusOK, rec.Code)
```

## 9. Refactoring Tips for Testability

1. **Depend on interfaces, not concrete types.** Extract small interfaces close to consumer side.
2. **Return (value, error)** instead of panicking – facilitates error assertions.
3. **Keep functions small & pure**; minimise global state.
4. **Inject clock / time.Now() function** to test expiry logic deterministically.
5. **Isolate randomness** – pass `rand.Rand` instance so you can seed.

## 10. Learning Resources

* Context7 – `/stretchr/testify` docs for assertions & mock examples.
* Context7 – `/golang/go` testing package guides.
* Blog posts (2023–2025):
  * "Unit Testing best practices in Golang" – Dwarves Foundation.
  * FOSSA "Best Practices for Testing in Go" (2021).
  * GRID Esports "Testing in Go" (2023).
* Official Go blog – [Table-driven tests](https://go.dev/blog/table-driven-tests).

## 11. Deliverables Checklist for Intern

- [ ] Local dev environment with `go1.22`, Make, Docker.
- [ ] Add Testify & GoMock to `go.mod`; commit generated mocks.
- [ ] Refactor auth & services for interface-driven design (if needed).
- [ ] Unit-tests covering **auth**, **config**, **handlers/health**, **services/embedding**.
- [ ] `make test-cover` passes ≥ 35 % by M2.
- [ ] Add CI workflow & badge to README.
- [ ] Documentation update: testing section in `README.md`.

---

Happy testing! 🎉  Reach out to the maintainers on Slack `#backend` channel for any blockers. 