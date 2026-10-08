# Agent notes

## Homebrew: always use the full cask name

The CLI ships as the cask `ghost-vk/tap/asana`. The bare name `asana` is the official Asana desktop app cask in homebrew/cask.

    brew install --cask ghost-vk/tap/asana
    brew upgrade --cask ghost-vk/tap/asana

Never run `brew upgrade asana` or `brew upgrade --cask asana`: brew resolves it to the desktop app, removes the CLI and takes over `/Applications/Asana.app`.

If it already happened: move `$(brew --prefix)/Caskroom/asana` aside (do not `brew uninstall` — that deletes the desktop app), then `brew install --cask ghost-vk/tap/asana`.

## Release

Tag `vX.Y.Z` on master → `.github/workflows/release.yml` runs goreleaser, publishes the release and updates the cask in `ghost-vk/homebrew-tap`. Bump `version` in `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json` when skills change.

## Testing against Asana

Do not touch real project data. Create a scratch task (`asana cr -a me "... (delete me)"`), exercise it, then `asana rm <gid>`.
