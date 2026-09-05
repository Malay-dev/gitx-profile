# Contributing to git-profile

Thanks for your interest in contributing! This project is part of the **gitx ecosystem** — a set of tools that add superpowers to Git.

## Getting Started

### Prerequisites

- Go 1.22+
- Git 2.36+

### Setup

```bash
git clone https://github.com/Malay-dev/gitx-profile.git
cd gitx-profile
go mod download
go build -o git-profile .
```

### Run Tests

```bash
go test ./... -v
```

### Test the binary

```bash
./git-profile --help
./git-profile doctor
```

## Project Structure

```
gitx-profile/
├── cmd/                    # CLI commands (cobra)
│   ├── root.go             # Root command
│   ├── add.go              # git profile add
│   ├── use.go              # git profile use
│   ├── list.go             # git profile list
│   ├── current.go          # git profile current
│   ├── show.go             # git profile show
│   ├── check.go            # git profile check
│   ├── remove.go           # git profile remove
│   ├── doctor.go           # git profile doctor
│   └── policy.go           # git profile policy
├── internal/
│   ├── profile/            # Profile storage, matching, types
│   ├── git/                # Git CLI wrapper
│   ├── config/             # File paths, env overrides
│   ├── policy/             # Policy evaluation engine
│   └── ui/                 # Colored terminal output
├── main.go                 # Entry point
├── plugin.yaml             # gitx ecosystem manifest
└── idea.md                 # Vision document
```

## Design Principles

1. **Never replace Git** — we augment it. `git commit`, `git push` etc. are untouched.
2. **Shell out to git** — we call `git config`, never edit `.git/` directly.
3. **Repository-scoped by default** — `git profile use work` writes to `.git/config`, not global.
4. **Transparent** — always show what changed and where.
5. **Zero runtime dependencies** — single binary, no Node/Python/Ruby needed.

## How to Contribute

### Bug Reports

Open an issue with:
- What you expected
- What happened
- Steps to reproduce
- `git profile doctor` output

### Feature Requests

Open an issue describing:
- The problem you're solving
- Your proposed solution
- Alternatives you considered

### Pull Requests

1. Fork the repo
2. Create a branch (`git checkout -b feat/my-feature`)
3. Write tests for new functionality
4. Ensure all tests pass (`go test ./...`)
5. Ensure code is formatted (`gofmt -w .`)
6. Commit with clear messages
7. Open a PR against `main`

### Code Style

- Follow standard Go conventions
- Run `go vet ./...` before submitting
- Run `gofmt` on all files
- Write table-driven tests where possible
- Keep functions focused — one function, one job

### Commit Messages

Use clear, descriptive messages:

```
feat: add --from flag to inherit profiles from global config
fix: handle missing ~/.gitprofiles gracefully
test: add policy evaluation tests for strict mode
docs: update README with show command
```

## Adding a New Command

1. Create `cmd/<command>.go`
2. Register it in `init()` with `rootCmd.AddCommand()`
3. Add business logic in `internal/` packages
4. Add tests in `internal/<package>/<file>_test.go`
5. Update README.md

## Questions?

Open an issue or start a discussion. We're building this in the open.
