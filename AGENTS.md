# AGENTS.md

This document provides guidelines for agentic coding agents working in this repository.

## Project Overview

storagebrowser v2 - A web-based file browser written in Go (backend) and Vue.js 3 + TypeScript (frontend). See `README.md` for full documentation.

---

## Build, Lint, and Test Commands

### Quick Reference

| Command | Description |
|---------|-------------|
| `make build` | Build binary (frontend + backend) |
| `make test` | Run all tests (frontend + backend) |
| `make lint` | Run all linters (frontend + backend) |
| `make fmt` | Format Go source files |

### Backend (Go)

| Command | Description |
|---------|-------------|
| `go build -o .` | Build backend binary |
| `go test -v ./...` | Run all backend tests |
| `go test -v ./package/...` | Run tests for specific package |
| `go test -run TestName -v` | Run single test by name |
| `./bin/golangci-lint run -v` | Run golangci-lint |
| `make fmt` | Format Go with goimports (local-prefix: github.com/futureharmony/storagebrowser) |

### Frontend (Vue.js + TypeScript)

Commands run from `frontend/` directory with `pnpm`:

| Command | Description |
|---------|-------------|
| `pnpm install` | Install dependencies (use `--frozen-lockfile` in CI) |
| `pnpm run build` | Build frontend (includes typecheck) |
| `pnpm run typecheck` | TypeScript type checking (vue-tsc) |
| `pnpm run lint` | Run ESLint |
| `pnpm run lint:fix` | Run ESLint with auto-fix |
| `pnpm run format` | Format with Prettier |
| `pnpm run test` | Run Playwright tests |

---

## Code Style Guidelines

### Go (Backend)

**Linting**: golangci-lint with `.golangci.yml` configuration

Key linters enabled:
- `goimports` - Import formatting (local-prefix: `github.com/futureharmony/storagebrowser`)
- `errcheck`, `errorlint` - Error handling
- `exhaustive` - Exhaustiveness checking for switch statements
- `govet` - Vet checks including `nilness` and `shadow`
- `staticcheck` - Static analysis
- `testifylint` - Test assertions

**Formatting**:
- Use `make fmt` (runs goimports with local prefix)
- Max line length: 140 characters (lll linter)
- Function max: 100 lines, 50 statements (funlen)
- Complexity threshold: 15 (gocyclo)

**Naming Conventions**:
- Use `PascalCase` for exported names, `camelCase` for unexported
- Use `CamelCase` for acronyms (e.g., `HTTPRequest`, not `HttpRequest`)
- Avoid `Foo` / `bar` naming; use descriptive names
- Receiver name: 1-letter like `r`, `c`, `s` for receiver methods

**Error Handling**:
- Use `errors.New()` for sentinel/static errors, `fmt.Errorf()` with `%w` for wrapping
- Define error variables in `errors/` package (see `errors/errors.go`)
- Use `errorlint` for proper error comparison
- Use `t.Fatalf()` in tests, not `t.Error()` followed by `t.FailNow()`

**Imports**:
- Standard library first, then third-party
- Group imports: stdlib | third-party | current module
- Use goimports auto-formatting (`make fmt`)

**Testing**:
- Use `t.Parallel()` for parallelizable tests
- Use table-driven tests with `t.Run()` subtests
- Use `t.Helper()` for helper functions
- Use testify assertions where appropriate

**Nolint Directives**:
- Must be specific: `//nolint:lintername`
- Require explanation if `nolintlint` requires it (config: `require-explanation: false`)
- Prefer fixing issues over adding nolint

### Frontend (Vue.js + TypeScript)

**Linting**: ESLint + Prettier with `eslint.config.js` and `prettier.config.js`

**Commands**:
- `pnpm run lint` - ESLint checks
- `pnpm run lint:fix` - Auto-fix issues
- `pnpm run format` - Prettier formatting

**TypeScript**:
- Strict mode enabled in `tsconfig.json`
- Use explicit types for function parameters and return values
- Avoid `any`; use `unknown` or proper types

**Vue 3 Composition API**:
- Use `<script setup>` syntax
- Use Composition API patterns consistent with existing code

---

## Project Conventions

### Directory Structure

```
/                       # Go backend root (module: github.com/futureharmony/storagebrowser/v2)
├── cmd/                # CLI commands (cobra)
├── errors/            # Error definitions
├── files/             # File handling
├── frontend/           # Vue.js frontend (separate npm package)
├── http/              # HTTP handlers
├── library/           # Custom libraries (e.g., afero-s3)
├── settings/          # Settings management
├── storage/           # Storage backends (bolt, etc.)
└── users/             # User management
```

### Go Module Path

Always use the full module path for imports:
```
github.com/futureharmony/storagebrowser/v2/package
```

### Environment Variables

- Configuration via `settings.local.json` (gitignored)
- See `settings.example.json` for template

### Commit Messages

- Uses commitlint (Conventional Commits)
- Run `make lint-commits` to validate

### Versioning

- Uses `standard-version` for versioning
- Run `make bump-version` to bump version

---

## Key Dependencies

- **Database**: Storm (BoltDB), bbolt
- **HTTP**: gorilla/mux, gorilla/websocket
- **Storage**: AWS S3 SDK v2, MinIO
- **CLI**: spf13/cobra, spf13/viper
- **Frontend**: Vue 3, Pinia, Vue Router, Vite

---

## IDE Recommendations

- **VS Code**: Use Go extension, Vue - Official, ESLint
- Configure `go.formatTool` to `goimports`
- Run `make fmt` on save or configure format on save

---

## Additional Resources

- `CONTRIBUTING.md` - Contribution guidelines
- `README.md` - Project documentation
- `.golangci.yml` - Full linter configuration
- `Makefile` - Build targets and commands
