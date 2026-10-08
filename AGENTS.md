# Agent notes

- Homebrew: only `ghost-vk/tap/asana` (`brew upgrade --cask ghost-vk/tap/asana`). Bare `asana` is the Asana desktop app.
- Release: tag `vX.Y.Z` on master. Bump `version` in `.claude-plugin/` and `.codex-plugin/plugin.json` when skills change.
- Live tests: scratch tasks only (`asana cr -a me "… (delete me)"`, then `asana rm`).
