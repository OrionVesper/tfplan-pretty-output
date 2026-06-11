# Contributing to tfplan-pretty-output

Thanks for your interest in contributing! 🎉

This document explains how to get started and how to make your contribution easy to review.

## 🎯 Ways to Contribute

- 🐛 **Report bugs** — open an issue with steps to reproduce and `doctor` output
- 💡 **Request features** — open an issue describing the problem you're solving
- 📝 **Improve documentation** — README, INSTALL, help text, FAQs
- 🧪 **Manual testing** — try edge cases, report what you find
- 🛠️ **Submit code** — PRs welcome (see workflow below)

## 🏗️ Development Setup

```bash
# 1. Fork the repo on GitHub

# 2. Clone your fork
git clone https://github.com/YOUR-USERNAME/tfplan-pretty-output.git
cd tfplan-pretty-output

# 3. Add upstream remote
git remote add upstream https://github.com/OrionVesper/tfplan-pretty-output.git

# 4. Install dependencies and build
go build ./...

# 5. Run doctor to verify your setup
go run . doctor
```

### Daily Workflow

```bash
# Sync with upstream
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feat/your-feature-name

# Build and test
go build ./...

# Install your local version
./install.sh

# Test the changes manually
tfplan-pretty-output plan test-feature
tfplan-pretty-output view test-feature

# Commit and push
git add .
git commit -m "feat: clear description of change"
git push origin feat/your-feature-name

# Open a PR on GitHub
```

## 📋 PR Workflow

1. **Open an issue first** for non-trivial changes — discuss the approach before coding
2. **Branch from `main`** — never commit directly to `main`
3. **Make focused commits** — one logical change per commit
4. **Test manually** — at minimum, run `tfplan-pretty-output doctor` and verify your change works
5. **Update CHANGELOG.md** — add an entry under `[Unreleased]`
6. **Open a PR** with a clear description of what changed and why

### Commit Message Convention

Use [Conventional Commits](https://www.conventionalcommits.org/) for clarity:

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation only
- `refactor:` code restructure (no behavior change)
- `chore:` build, tooling, dependencies
- `test:` adding or improving tests

Examples:
```
feat: add --color flag to show command
fix: handle empty plan output correctly
docs: clarify INSTALL.md for Linux ARM64
```

## 🎨 Code Style

### Go Conventions

- Run `gofmt` before committing — most editors do this automatically
- Use meaningful names — avoid `tmp`, `data`, `obj` without context
- Keep functions focused — split anything over ~50 lines
- Comment exported functions (`// FunctionName ...`)
- Return errors, don't panic
- Prefer explicit error handling over deferred handling

### Project Conventions

- **Mandatory plan names** — never reintroduce sticky-name memory
- **No config files** — defaults are intentional; ask before adding configurability
- **Lineage immutability** — once a plan is saved, never modify it (only delete)
- **Custom errors** over Cobra's default error output
- **Brand rewriting** — passthrough commands rewrite `terraform` → `tfplan-pretty-output`

### What NOT to Add

These were intentionally removed or excluded. **Open an issue to discuss before reintroducing:**

- ❌ `-out` flag handling (we manage plan files internally)
- ❌ Sticky-name memory (users must always specify the name)
- ❌ Hash-based plan IDs (we use monotonic numbers)
- ❌ Prompts before plan capture (users provide name on the command line)
- ❌ Vim-mode keys (deferred until requested)
- ❌ Mouse support (deferred until requested)
- ❌ YAML or other config files (we removed this — defaults are fine)

## 🧪 Testing

Automated tests are not yet in place. Until they are, **every PR must include manual testing**.

In your PR description, describe what you tested. At minimum:

- ✅ `make build` (or `go build ./...`) succeeds
- ✅ `tfplan-pretty-output doctor` reports all green
- ✅ Smoke test relevant scenarios for your change

### Smoke Test (5 min)

```bash
make install
mkdir -p /tmp/tfsandbox && cd /tmp/tfsandbox
cat > main.tf <<'EOF'
resource "null_resource" "demo" {}
EOF
terraform init

tfplan-pretty-output plan smoketest
tfplan-pretty-output view smoketest
tfplan-pretty-output remove smoketest -y
```

## 🐛 Reporting Bugs

Please include:

- **Steps to reproduce** — be specific
- **Output of `tfplan-pretty-output doctor`** — full output
- **Terraform version** — `terraform version`
- **OS** — macOS or Linux + version
- **Expected vs actual behavior**

## 💡 Requesting Features

Please include:

- **The problem you're solving** — why this is needed
- **Your proposed solution** — what the feature would look like
- **Alternatives considered** — why existing features don't suffice

## 📜 License

By contributing, you agree your contributions will be licensed under the project's [MIT License](LICENSE).

## 🙏 Thanks!

Every contribution helps, whether it's code, docs, bug reports, or just trying the tool and giving feedback. 🚀