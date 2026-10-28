# Repository Guidelines

## Project Structure & Module Organization

This repository contains four Go modules; there is no root `go.mod` or `go.work`.

- `ai/ai-agent-demo/`: direct LLM example and shared `.env` loader in `env/`.
- `ai/ai-agent-demo/framework_agent/`: nested LangChainGo agent module; its `go.mod` references the parent module through a local `replace` directive.
- `ai/rag-test/`: RAG HTTP server, with `internal/api`, `config`, `parser`, `rag`, and `storage` packages. PDFs and usage documentation live alongside `main.go`.
- `sse-demo/`: SSE server and browser client in `static/index.html`.

## Build, Test, and Development Commands

Use Go 1.24.4 or newer for all modules. Run commands from the relevant module directory:

```bash
go build ./...       # Compile packages
go test ./...        # Run module tests
go vet ./...         # Check common Go mistakes
gofmt -w path/to/changed.go
```

Repeat checks separately for the nested agent module; parent-module commands do not include it.

- `cd ai-agent-demo && go run .`: run the direct LLM demo.
- `cd ai/ai-agent-demo/framework_agent && go run .`: run the framework agent.
- `cd sse-demo && go run .`: serve the browser client on port 8080; check streaming with `curl -N http://localhost:8080/events`.
- `cd rag-test && go run . -port 8080`: start RAG; requires `pdftotext`, Ollama models, and Qdrant. Check `/api/health`. Run servers on separate ports when needed.

## Coding Style & Naming Conventions

Use `gofmt` formatting and tabs for Go indentation. Keep package names lowercase, exported identifiers in PascalCase, and unexported identifiers in camelCase. Keep HTTP handling, configuration, parsing, and storage in their existing packages. Wrap errors with useful context.

## Testing Guidelines

No `*_test.go` files or coverage threshold currently exist. Add standard-library `testing` tests beside the affected code, named `TestXxx` in `*_test.go`. Prefer table-driven cases and `httptest` for HTTP behavior. Mock external services for deterministic unit tests; document prerequisites for integration checks.

## Commit & Pull Request Guidelines

History currently contains only `demo`; no formal commit convention is established. Write concise, imperative subjects identifying the affected demo. PRs should describe behavior changes, list validation commands and results, link relevant issues, and include screenshots for browser UI changes.

## Security & Agent Guidance

Copy `ai/ai-agent-demo/.env.example` to `.env` for local credentials. Never commit secrets, personal absolute paths, or private service addresses. Use placeholders in documentation.

When available, use CodeGraph for structural questions and native search for literal text. Prefer `codegraph_context` followed by one `codegraph_explore` call. If the server reports an uninitialized index, ask before running `codegraph init -i`.
