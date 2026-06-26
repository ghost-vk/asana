---
name: asana-install
description: Install the asana CLI.
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

It prints the URL to create a token (Asana → Settings → Apps → Personal Access Tokens), then prompts you to paste it and pick a workspace. The CLI reads config from `~/.config/asana-cli/config.yml` first, with `~/.asana.yml` as a fallback; `asana config` writes `~/.asana.yml`.
