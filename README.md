<p align="center">
  <img src="gitx-profile.png" alt="git-profile"/>
</p>

<h1 align="center">git-profile</h1>

<p align="center">
  <a href="https://github.com/Malay-dev/gitx-profile/actions/workflows/ci.yml"><img src="https://github.com/Malay-dev/gitx-profile/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://goreportcard.com/report/github.com/Malay-dev/gitx-profile"><img src="https://goreportcard.com/badge/github.com/Malay-dev/gitx-profile" alt="Go Report Card" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT" /></a>
</p>

<p align="center"><strong>Manage multiple Git identities. Never commit with the wrong email again.</strong></p>

```
$ git profile use Personal

  ✓ Switched to profile "Personal"

    Scope:             Local
    Repository:        my-side-project
    Name:              Malay Kumar
    Email:             malay@gmail.com
```

`git-profile` is a Git extension that lets you define named profiles (work, personal, open-source) and switch between them per repository. It integrates with Git's native extension system — just type `git profile`.

Part of the [gitx](https://github.com/Malay-dev) ecosystem.

## The Problem

If you use Git with multiple accounts, you've probably:

- Committed to a work repo with your personal email
- Had to `git config user.email ...` in every new clone
- Forgotten which identity a repository is using
- Wished Git had a built-in "profile" concept

`git-profile` fixes all of this.

## Quick Start

```bash
# Install
go install github.com/Malay-dev/gitx-profile@latest

# Create profiles
git profile add personal --name "Alice" --email "alice@gmail.com"
git profile add work --name "Alice" --email "alice@acmecorp.com"

# Or inherit from your existing global config
git profile add work --from global

# Switch identity (current repo only)
git profile use work

# See what's active
git profile current

# Inspect a repo's full identity setup
git profile show
```

## Install

### From source (requires Go 1.22+)

```bash
go install github.com/Malay-dev/gitx-profile@latest
```

The binary is named `git-profile`. Git automatically discovers it as `git profile`.

### From GitHub Releases

Download the binary for your platform from [Releases](https://github.com/Malay-dev/gitx-profile/releases), extract it, and place it somewhere in your `PATH`:

```bash
# macOS / Linux
sudo mv git-profile /usr/local/bin/

# Or without sudo
mkdir -p ~/.local/bin
mv git-profile ~/.local/bin/
```

### Verify installation

```bash
git profile --help
```

If Git finds it, you're good.

## Commands

### `git profile add <name>`

Create a new profile.

```bash
# Interactive
git profile add work

# With flags
git profile add work --name "Alice" --email "alice@acmecorp.com"

# Inherit from global git config (with confirmation prompt)
git profile add work --from global

# Clone from an existing profile (override what's different)
git profile add oss --from work --email "alice@opensource.org"
```

### `git profile use <name>`

Switch to a profile. **Writes to the current repo by default** (not global).

```bash
git profile use work
# ✓ Switched to profile "work"
#
#   Scope:             Local
#   Repository:        inventory-service
#   Name:              Alice
#   Email:             alice@acmecorp.com
```

Use `--global` to change the global default:

```bash
git profile use personal --global
```

### `git profile current`

Show the active identity and which profile it matches.

```bash
git profile current
# ✓ Profile: work
#
#   Name:              Alice
#   Email:             alice@acmecorp.com
#   Repository:        inventory-service
#   Scope:             Local (.git/config)
```

### `git profile list`

List all saved profiles. Active one is marked with `*`.

```bash
git profile list
# * work <alice@acmecorp.com>
#   personal <alice@gmail.com>
#   oss <alice@opensource.org>
```

### `git profile show`

Discover the full identity configuration of the current repository.

```bash
git profile show
# === Repository ===
#
#   Name:              inventory-service
#   Remote (origin):   git@github.com:AcmeCorp/inventory-service.git
#   Branch:            main
#
# === Identity ===
#
#   user.name:         Alice
#   user.email:        alice@acmecorp.com
#
#   Source breakdown:
#     user.name:       Alice (local — .git/config)
#     user.email:      alice@acmecorp.com (local — .git/config)
#
# === Profile ===
#
#   ✓ Matched: work
```

### `git profile check`

Verify the current identity matches a known profile. Useful in pre-commit hooks.

```bash
git profile check
# ✓ Identity matches profile "work"

git profile check --quiet
# (exit code 0 = match, 1 = mismatch)
```

### `git profile doctor`

Run diagnostics on your Git identity setup.

```bash
git profile doctor
```

### `git profile remove <name>`

Delete a saved profile. Does not change any existing git config.

```bash
git profile remove old-work
```

### `git profile policy list` / `git profile policy create`

Manage identity policies — control which profiles are allowed in which repos.

```bash
git profile policy create
# Policy name: acme-corp
# Remote patterns: github.com/AcmeCorp/*
# Allowed profiles: work
# Mode (strict/warning/advisory): strict
#
# ✓ Policy "acme-corp" created
```

## How It Works

Profiles are stored in `~/.gitprofiles`:

```ini
[profile "personal"]
  name = Alice
  email = alice@gmail.com

[profile "work"]
  name = Alice
  email = alice@acmecorp.com
  signingkey = ABC123
```

When you run `git profile use work`, the tool executes:

```bash
git config user.name "Alice"
git config user.email "alice@acmecorp.com"
```

That's it. No magic. Git remains the source of truth.

### Scope

```
git profile use work           → .git/config (this repo only)
git profile use work --global  → ~/.gitconfig (all repos without local override)
```

### Policies

Policies prevent using the wrong profile in a repo. Two levels:

**User-level** (`~/.git-profile/policies.yaml`):
```yaml
policies:
  - name: acme-corp
    remotePatterns:
      - "github.com/AcmeCorp/*"
    allowedProfiles:
      - work
    mode: strict
```

**Repo-level** (`.gitprofile.yml` in repo root):
```yaml
organization: AcmeCorp
allowedProfiles:
  - work
mode: strict
```

Modes: `strict` (blocks), `warning` (asks), `advisory` (suggests).

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GITX_PROFILES_PATH` | Custom path to profiles file | `~/.gitprofiles` |
| `GITX_POLICIES_PATH` | Custom path to policies file | `~/.git-profile/policies.yaml` |

## Roadmap

- [x] Core CLI (add, list, use, current, show, check, remove, doctor)
- [x] `--from` flag (inherit from global config or existing profile)
- [x] Policy engine (user-level + repo-level)
- [x] CI/CD (GitHub Actions + GoReleaser)
- [ ] Remote-based auto-suggestions
- [ ] Pre-commit hook integration (`git profile enable-guard`)
- [ ] SSH key switching (v2)
- [ ] gitx plugin marketplace (separate project)

## Part of the gitx Ecosystem

`git-profile` is the first tool in a broader vision: **gitx** — a platform for Git extensions.

Future plugins:
- `git-guard` — enforce branch and commit policies
- `git-doctor` — comprehensive Git health diagnostics
- `git-secrets` — prevent secrets from being committed

Each plugin is standalone, installable via `gitx install`, and follows Git's native extension pattern.

See [idea.md](idea.md) for the full vision.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup instructions, project structure, and guidelines.

## License

[MIT](LICENSE)
