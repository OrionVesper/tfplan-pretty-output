<h1 align="center">tfplan-pretty-output</h1>

<p align="center">
  <strong>Terminal-native, Terraform plan reviewer with persistent named snapshots and an interactive viewer.</strong>
</p>

---

## 📖 Table of Contents

- [About](#-about)
- [The Problem](#-the-problem)
- [The Solution](#-the-solution)
- [Features](#-features)
- [Demo](#-demo)
- [Installation](#-installation)
  - [Pre-built Binaries (Recommended)](#pre-built-binaries-recommended)
  - [Build From Source (Developers)](#build-from-source-developers)
- [Usage](#-usage)
  - [Quick Start](#quick-start)
  - [Commands](#commands)
  - [TUI Key Bindings](#tui-key-bindings)
  - [Examples](#examples)
- [Architecture](#-architecture)
- [FAQ](#-faq)
- [Contributing](#-contributing)
- [License](#-license)
- [Acknowledgments](#-acknowledgments)

---

## 🎯 About

`tfplan-pretty-output` is a command-line tool that wraps the Terraform plan/apply lifecycle with **persistent named snapshots** and an **interactive terminal-based viewer**. It's designed for engineers who review Terraform plans daily and want a faster, safer, and more readable workflow.

The tool acts as a transparent wrapper around `terraform` — every standard terraform command (`init`, `validate`, `fmt`, `apply`, `destroy`, etc.) works as expected, while `plan` and `apply` get supercharged with lineage-based storage and a rich TUI.

---

## 🐛 The Problem

When you run `terraform plan`, the output is overwhelming:

- A wall of text scrolls through your terminal in seconds
- Critical destroys (`-`) are buried among unchanged context lines
- Force-replacements (`±`) are easy to miss
- There's no way to compare today's plan with yesterday's
- The plan is lost the moment you close your terminal
- Reviewing 100+ resource changes requires careful scrolling — and humans miss things
- Applying blindly trusts whatever's in your terminal scrollback

Real-world consequences:

- ❌ Engineers accidentally approve destructive changes they didn't notice
- ❌ Drift between code and infrastructure goes undetected
- ❌ Code review of infrastructure changes is impractical because plans aren't saved
- ❌ Auditing what changed (and when) requires scrolling through Slack screenshots

This tool exists because **plan review deserves better than `cat` and prayer**.

---

## ✨ The Solution

`tfplan-pretty-output` transforms the Terraform workflow:

1. **Plans are persistent and named.** Every plan is saved with a lineage name you choose:
   ```bash
   tfplan-pretty-output plan production
   tfplan-pretty-output plan staging
   ```

2. **Plans get an interactive viewer.** Press one command and review them with keyboard shortcuts:
   ```bash
   tfplan-pretty-output view production
   ```

3. **Apply is safer.** It only applies the latest plan in a lineage, refusing older plans that may have drifted from current state:
   ```bash
   tfplan-pretty-output apply production
   ```

4. **Storage is automatic.** Each lineage keeps the 3 most recent plans (older are pruned). Each project keeps the 3 most recently used lineages (LRU evicted). No config required.

5. **Everything else passes through transparently.** All standard Terraform commands behave exactly like raw `terraform` — your existing workflows just work.

---

## ✨ Features

- 🎯 **Mandatory named plans** — no anonymous snapshots cluttering your workspace
- 🔄 **Lineage-based history** — 3 plans per lineage, 3 lineages per project, all auto-managed
- 🎨 **Interactive TUI** — navigate with arrow keys, expand resources inline, search across the full plan
- 🌈 **Verbose mode** — view the entire plan with native Terraform colors, top-to-bottom
- 🔍 **Smart search** — `/text` searches resource addresses and body content; matches highlighted with `n`/`N` cycling
- 🛡️ **Apply-only-latest** — prevents the "apply old plan, corrupt state" footgun
- 🩺 **Built-in `doctor`** — diagnose setup issues with one command
- 🚪 **Graceful Ctrl+C** — terraform subprocess cleanup, no orphans
- 🔗 **Workdir symlinks** — `./production` always points to the latest plan for that lineage
- 🪶 **Transparent wrapper** — all 11 terraform subcommands pass through correctly
- 📦 **Single binary** — zero dependencies, instant install

---

## 🎬 Demo

📺 *Demo video coming soon* — link will be added to this README.

Here's what a typical session looks like:

```
$ tfplan-pretty-output plan production
[terraform plan runs...]
✓ Plan captured: production-1

To view, run:
  tfplan-pretty-output view production

To apply, run:
  tfplan-pretty-output apply production
```

The interactive viewer (`tfplan-pretty-output view production`):

- List of resources on the left, colored symbols on the right
- Cursor navigates with `↑↓`
- `Enter` expands a resource to see its diff inline
- `v` switches to verbose mode (full plan, terraform colors)
- `/` searches across the plan
- `?` shows all keyboard shortcuts

---

## 📦 Installation

### Pre-built Binaries (Recommended)

Pre-built binaries are available for macOS (Intel + Apple Silicon) and Linux (x64 + ARM64).

**macOS (Apple Silicon — M1/M2/M3/M4):**

```bash
curl -L https://github.com/OrionVesper/tfplan-pretty-output/releases/latest/download/tfplan-pretty-output_Darwin_arm64.tar.gz | tar xz
sudo mv tfplan-pretty-output /usr/local/bin/
tfplan-pretty-output doctor
```

**macOS (Intel):**

```bash
curl -L https://github.com/OrionVesper/tfplan-pretty-output/releases/latest/download/tfplan-pretty-output_Darwin_x86_64.tar.gz | tar xz
sudo mv tfplan-pretty-output /usr/local/bin/
tfplan-pretty-output doctor
```

**Linux (x64):**

```bash
curl -L https://github.com/OrionVesper/tfplan-pretty-output/releases/latest/download/tfplan-pretty-output_Linux_x86_64.tar.gz | tar xz
sudo mv tfplan-pretty-output /usr/local/bin/
tfplan-pretty-output doctor
```

**Linux (ARM64):**

```bash
curl -L https://github.com/OrionVesper/tfplan-pretty-output/releases/latest/download/tfplan-pretty-output_Linux_arm64.tar.gz | tar xz
sudo mv tfplan-pretty-output /usr/local/bin/
tfplan-pretty-output doctor
```

**Windows:** Native Windows is not supported (the tool uses Unix process groups and symlinks). Use [WSL2](https://learn.microsoft.com/en-us/windows/wsl/install) and follow the Linux instructions above.

### Build From Source (Developers)

If you want to build from source — for contributions, custom modifications, or to use a development version:

**Prerequisites:**

- Go 1.21 or later ([go.dev/dl](https://go.dev/dl/))
- Git

```bash
# Clone the repository
git clone https://github.com/OrionVesper/tfplan-pretty-output.git
cd tfplan-pretty-output

# Install
./install.sh

# Verify
tfplan-pretty-output doctor
```

The `install.sh` script runs `go install ./...`, which installs the binary to `$(go env GOPATH)/bin`. Make sure that directory is on your `PATH`:

```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

For full installation details, see [INSTALL.md](INSTALL.md).

---

## 🚀 Usage

### Quick Start

```bash
# 1. Go to your terraform project
cd ~/your-terraform-project

# 2. Capture a plan with a name
tfplan-pretty-output plan production

# 3. Review it in the interactive viewer
tfplan-pretty-output view production

# 4. Apply when ready
tfplan-pretty-output apply production
```

### Commands

#### Terraform Lifecycle Commands.

| Command | Description |
|---------|-------------|
| `init` | Initialize the Terraform working directory |
| `validate` | Check configuration syntax |
| `fmt` | Rewrite configuration files to canonical format |
| `plan <name>` | Run terraform plan and save snapshot |
| `apply <name>` | Apply the latest plan in a lineage |
| `destroy` | Destroy all remote objects |
| `force-unlock <id>` | Release a stuck state lock |

#### Terraform Inspection Commands.

| Command | Description |
|---------|-------------|
| `output` | Show terraform output values |
| `refresh` | Update state to match real infrastructure |
| `providers` | Show required providers |

#### tfplan-pretty-output Specific Commands

| Command | Description |
|---------|-------------|
| `view <name>` | Open the latest plan in the interactive TUI |
| `view <name>-<N>` | Open a specific older plan |
| `view list` | List all saved plans for the current project |
| `diff <id1> <id2>` | Compare two plans, show added/removed/changed resources |
| `show <id>` | Print plan as pipeable text |
| `remove <id\|name\|--all>` | Delete saved plans (with `-y` to skip confirmation) |
| `auto-clean` | Don't create or save snapshots. Clean run for CI/CD workflow |
| `doctor` | Run health checks on your setup |
| `version` | Show tfplan-pretty-output and terraform versions |
| `help [command]` | Get help (overview or detailed per-command) |

Run `tfplan-pretty-output help <command>` for detailed documentation on any command.

### TUI Key Bindings

When the interactive viewer is open:

**Navigation:**

- `↑` `↓`      — Move between resources
- `g` `G`      — Jump to top / bottom
- `Enter`      — Expand / collapse the current resource
- `Shift + ↑↓` — Scroll within an expansion
- `Esc`        — Close expansion / clear search / quit
- `q`          — Quit

**Modes:**

- `v` — Switch to verbose mode (all resources expanded, scrollable)
- `?` — Show keyboard help overlay

**Search:**

- `/<text>` — Search across resource addresses and body content (matches highlighted)
- `/<number>` — Jump to a specific row and auto-expand
- `n` — Next match
- `N` — Previous match

### Examples

**Capture a plan and open the viewer in one command:**

```bash
tfplan-pretty-output plan production -- view
```

**Run a plan with extra terraform flags:**

```bash
tfplan-pretty-output plan production -refresh-only -target=aws_instance.web
```

**Compare yesterday's plan with today's:**

```bash
tfplan-pretty-output diff production-1 production-2
```

**Pipe a plan to other tools:**

```bash
tfplan-pretty-output show production | grep "destroy"
tfplan-pretty-output show production > plan-audit-$(date +%Y%m%d).txt
```

**Capture a plan for CI/CD without saving:**

```bash
tfplan-pretty-output plan ci -- auto-clean
```

**Clean up a specific lineage:**

```bash
tfplan-pretty-output remove production -y
```

**Clean up everything for the current project:**

```bash
tfplan-pretty-output remove --all -y
```

**Diagnose setup issues:**

```bash
tfplan-pretty-output doctor
```

---

## 🏗️ Architecture

Plans are stored under `~/.tfplan-pretty-output/projects/<project_id>/<lineage>/`. Each project's identity is tracked via a marker file in the workdir, so renaming or moving the directory doesn't break history.

**Storage layout:**

```
~/.tfplan-pretty-output/projects/<project_id>/
├── production/
│   ├── production-97.tfplan + production-97.json
│   ├── production-98.tfplan + production-98.json
│   └── production-99.tfplan + production-99.json    ← latest
├── staging/
└── test/

(working directory)
./production  → projects/<id>/production/production-99.tfplan  (symlink)
```

**Numbering:** Plans within a lineage are numbered monotonically. The 3 highest-numbered plans are kept; older plans are pruned automatically.

**Lineage limits:** Each project keeps the 3 most-recently-used lineages. When a 4th lineage is created, the least recently used is evicted entirely.

**Why this design:**

- ✅ Mandatory naming prevents anonymous plans
- ✅ Monotonic numbering makes plan IDs predictable (no hash gobbledygook)
- ✅ Auto-pruning prevents storage bloat
- ✅ Workdir symlinks let you reference plans naturally (`./production` always = latest)
- ✅ Project identity tracking means renaming workdirs doesn't break history
- ✅ No config files (deliberately minimal surface area)

---

## ❓ FAQ

**Why must I provide a name on every `plan`?**

To prevent anonymous plans from cluttering your storage and to make `apply` and `view` predictable. The name is your organizing principle — environment, branch, experiment, whatever fits your workflow.

**Why doesn't `apply <plan-id>` work for specific old plans?**

Old plans drift from current code. Applying a 2-week-old plan against today's terraform state is a great way to corrupt your infrastructure. To prevent this footgun, `apply` only works on the latest plan in a lineage. If you genuinely need to apply an older plan, use raw terraform.

**Can I use this with terraform versions earlier than 1.0?**

Tested with 1.0 and later. Earlier versions may work but are not officially supported.

**Does this work on Windows?**

Use [WSL2 (Windows Subsystem for Linux)](https://learn.microsoft.com/en-us/windows/wsl/install). The tool uses Unix-specific features (signal handling, process groups, symlinks) that don't translate cleanly to native Windows.

**Where can I report bugs or suggest features?**

Please open an issue at [GitHub Issues](https://github.com/OrionVesper/tfplan-pretty-output/issues). Include the output of `tfplan-pretty-output doctor` in bug reports.

**Is this affiliated with HashiCorp?**

No. This is an independent open-source project that wraps Terraform.

---

## 🤝 Contributing

Contributions are welcome. Please open an issue first for any non-trivial changes so we can discuss the approach.

For bug reports, please include the output of `tfplan-pretty-output doctor` in your issue.

---

## 📜 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

This project is built with the following excellent open-source libraries:

- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [BubbleTea](https://github.com/charmbracelet/bubbletea) — Modern TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [Go](https://go.dev/) — Programming language
- [GoReleaser](https://goreleaser.com/) — Release automation

Inspired by the daily frustration of reading Terraform plan output.