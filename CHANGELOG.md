# Changelog

All notable changes to this project will be documented in this file.

The format is based on https://keepachangelog.com/en/1.1.0/,
and this project adheres to https://semver.org/spec/v2.0.0.html.

## [Unreleased]

Future improvements will be tracked here.

## [0.1.0] - Initial Release

### Added

**Lifecycle commands (transparent terraform passthroughs):**
- `init` — Initialize Terraform working directory
- `validate` — Check configuration syntax
- `fmt` — Format configuration files
- `destroy` — Destroy infrastructure
- `force-unlock` — Release stuck state locks

**Custom commands:**
- `plan <name>` — Capture plan with mandatory lineage naming
- `apply <name>` — Apply latest plan (rejects older IDs for safety)
- `view <name | name-N | list>` — Open interactive TUI or list saved plans
- `diff <id1> <id2>` — Compare two plans
- `show <id>` — Pipeable text output
- `remove <id | name | --all>` — Delete saved plans with `-y` for scripts
- `doctor` — Run health checks
- `version` — Show tfplan-pretty-output and terraform versions
- `help [command]` — Overview or per-command help

**Storage:**
- Lineage-based monotonic numbering (e.g., `prod-1`, `prod-99`)
- 3 plans kept per lineage, oldest auto-pruned
- 3 lineages kept per project, LRU evicted
- Project ID survives directory renames
- Workdir symlinks always point to the latest plan

**TUI viewer:**
- Color-coded symbols (+, -, ~, ±) at right edge of each row
- Inline expansion with `Enter`
- Adaptive 20-line viewport
- Sticky cursor positioning
- Multi-line wrapping for long resource addresses
- Plan summary in header
- Output changes as a virtual row
- Serial numbers and `g`/`G` for top/bottom jumps
- `/<text>` search across address + body content
- `n`/`N` for match cycling
- Verbose mode (`v` key)
- Full-screen `?` help overlay

**Safety:**
- Mandatory plan names (no anonymous plans)
- Rejects `-out=` flag (we manage plan files)
- Apply rejects specific old plan IDs
- Defensive guards on missing files
- Custom error messages
- Graceful Ctrl+C with subprocess cleanup
- No orphan terraform processes

**Other:**
- Native terraform colors preserved in TUI
- Brand rewriting in passthrough output
- Pre-built binaries for macOS (Intel + Apple Silicon) and Linux (x64 + ARM64)
- MIT License