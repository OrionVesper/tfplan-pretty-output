# Installation Guide

## 📋 Prerequisites

- **Terraform 1.0 or later** — https://developer.hashicorp.com/terraform/install

That's it. The tool brings everything else it needs.

Verify your terraform install:

```bash
terraform version
```

---

## 📦 Installation Methods

There are two ways to install `tfplan-pretty-output`:

1. **Pre-built binaries** (recommended)
2. **From source** (for developers and contributors)

---

## 🚀 Option 1: Pre-built Binaries (Recommended)

Download the latest binary for your platform from the [Releases page](https://github.com/OrionVesper/tfplan-pretty-output/releases).

### macOS (arm64)

```bash
curl -L https://github.com/OrionVesper/tfplan-pretty-output/releases/latest/download/tfplan-pretty-output_Darwin_arm64.tar.gz | tar xz
sudo mv tfplan-pretty-output /usr/local/bin/
tfplan-pretty-output doctor
```

### macOS (x86_64)

```bash
curl -L https://github.com/OrionVesper/tfplan-pretty-output/releases/latest/download/tfplan-pretty-output_Darwin_x86_64.tar.gz | tar xz
sudo mv tfplan-pretty-output /usr/local/bin/
tfplan-pretty-output doctor
```

### Linux (x64)

```bash
curl -L https://github.com/OrionVesper/tfplan-pretty-output/releases/latest/download/tfplan-pretty-output_Linux_x86_64.tar.gz | tar xz
sudo mv tfplan-pretty-output /usr/local/bin/
tfplan-pretty-output doctor
```

### Linux (ARM64)

```bash
curl -L https://github.com/OrionVesper/tfplan-pretty-output/releases/latest/download/tfplan-pretty-output_Linux_arm64.tar.gz | tar xz
sudo mv tfplan-pretty-output /usr/local/bin/
tfplan-pretty-output doctor
```

### Windows

Native Windows is not supported (we use Unix process groups and symlinks). Use [WSL2](https://learn.microsoft.com/en-us/windows/wsl/install) and follow the Linux instructions above.

### Verify

```bash
tfplan-pretty-output version
# Should print version, commit, build date, plus terraform's version

tfplan-pretty-output doctor
# Should report: ✨ All checks passed
```

---

## 🛠️ Option 2: Build From Source (Developers)

If you have Go 1.21+ installed and want the latest development version:

### Prerequisites

- Go 1.21 or later — [go.dev/dl](https://go.dev/dl/)
- Git

### Install

```bash
git clone https://github.com/OrionVesper/tfplan-pretty-output.git
cd tfplan-pretty-output
./install.sh
```

This builds and installs the binary to `$(go env GOPATH)/bin`. Make sure that directory is on your `PATH`:

```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

---

## 🔄 Updating

### Binary installations

Re-run the curl command from Option 1 above. The new binary overwrites the old one.

### From source

```bash
cd tfplan-pretty-output
git pull
./install.sh
```

---

## 🗑️ Uninstalling

### Remove the binary

```bash
sudo rm /usr/local/bin/tfplan-pretty-output
# Or if installed from source:
rm $(go env GOPATH)/bin/tfplan-pretty-output
```

### Remove all saved plans

```bash
rm -rf ~/.tfplan-pretty-output
```

### Remove workdir symlinks

```bash
# In each project where you used tfplan-pretty-output plan
cd ~/your-terraform-project
ls -la                       # Find symlinks (lineage names)
rm production staging test   # Remove any you created
```

---

## 🩺 Troubleshooting

### `command not found: tfplan-pretty-output`

The binary isn't on your `PATH`. Add the install directory to your shell profile:

```bash
# For binary installs:
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.zshrc

# For source installs:
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc

source ~/.zshrc
```

### `terraform binary not found on PATH`

Install terraform first: https://developer.hashicorp.com/terraform/install

### Permission denied during install

The `sudo mv` command requires admin privileges. If you don't have sudo access, install to your user-local bin:

```bash
mkdir -p ~/.local/bin
mv tfplan-pretty-output ~/.local/bin/
echo 'export PATH=$PATH:~/.local/bin' >> ~/.zshrc
source ~/.zshrc
```

### Build errors from source

Try cleaning the module cache:

```bash
go clean -modcache
./install.sh
```

### For other issues

Run the built-in diagnostic:

```bash
tfplan-pretty-output doctor
```

It reports common setup problems with suggested fixes.

---

## 🆘 Still Stuck?

- Open an [issue on GitHub](https://github.com/OrionVesper/tfplan-pretty-output/issues)
- Check the [README FAQ](README.md#-faq)
- Run `tfplan-pretty-output doctor` and include the output in your issue