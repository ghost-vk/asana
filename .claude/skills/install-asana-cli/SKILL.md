---
name: install-asana-cli
description: Install the asana CLI (github.com/ghost-vk/asana), a command-line Asana client, onto this machine. Use whenever the user wants to install, set up, get, or update the `asana` command — including phrasings like "install asana cli", "get the asana command", "set up asana on my mac", "how do I install asana", or when `asana` is reported as command-not-found and they want it.
---

# Install asana CLI

Installs `asana` from `github.com/ghost-vk/asana`. Pick the method by platform; verify at the end.

## Choose a method

1. **macOS — Homebrew (preferred):**
   ```bash
   brew install ghost-vk/tap/asana
   ```
   This is a cask (prebuilt binary), macOS only. To update later: `brew upgrade asana`.

2. **Any OS with a Go toolchain:**
   ```bash
   go install github.com/ghost-vk/asana@latest
   ```
   Installs to `$(go env GOBIN)` or `$(go env GOPATH)/bin` — make sure that dir is on `$PATH`.

3. **Linux / no Go — prebuilt binary from a release:**
   Download the `.tar.gz` matching your OS+arch (darwin/linux, amd64/arm64) from
   https://github.com/ghost-vk/asana/releases/latest then:
   ```bash
   tar -xzf asana_*.tar.gz asana && sudo mv asana /usr/local/bin/
   ```

4. **From source (have the repo cloned):**
   ```bash
   go build -o asana . && sudo mv asana /usr/local/bin/
   ```

## Verify

```bash
asana --version
```
Should print `asana version X.Y.Z`. If "command not found", the install dir isn't on `$PATH`.

## First-time setup

The CLI needs an Asana Personal Access Token:
```bash
asana config
```
It prints the URL to create a token (Asana → Settings → Apps → Personal Access Tokens), then prompts you to paste it and pick a workspace. Config is saved to `~/.asana.yml`.
