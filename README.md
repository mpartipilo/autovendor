# autovendor

Automatic `go mod vendor` after git operations. Never run it manually again.

## The problem

If your Go project uses vendoring, every `git pull`, branch switch, or rebase that changes `go.mod` leaves `vendor/` out of sync. Your IDE breaks with errors like:

```text
Error loading workspace: packages.Load error: go: inconsistent vendoring
```

You have to remember to run `go mod vendor` every time. The Go team has [known about this since 2021](https://github.com/golang/go/issues/45161), but no upstream fix exists.

**autovendor** installs lightweight git hooks that detect dependency changes and automatically sync your vendor directory. It supports monorepos with multiple Go modules.

## Install

**Homebrew:**

```sh
brew install mpartipilo/autovendor/autovendor
```

Or tap once, then use the short name forever:

```sh
brew tap mpartipilo/autovendor
brew install autovendor
```

**Go install:**

```sh
go install github.com/mpartipilo/autovendor@latest
```

**mise:**

```sh
mise use -g github:mpartipilo/autovendor
```

**Quick try (no install):**

```sh
go run github.com/mpartipilo/autovendor@latest install
```

## Usage

### Set up hooks in a repository

```sh
cd ~/src/your-go-project
autovendor install
```

That's it. From now on, `go mod vendor` runs automatically after:

- **`git pull`** / `git merge` (post-merge hook)
- **`git checkout`** / branch switches (post-checkout hook)
- **`git rebase`** (post-rewrite hook)

It only runs when `go.mod` or `go.sum` actually changed, so there's no overhead on unrelated pulls.

### Install in a specific directory

```sh
autovendor install ~/src/another-project
```

### Remove hooks

```sh
autovendor uninstall
```

## How it works

1. `autovendor install` detects your repo's hooks directory (respects `core.hooksPath` config) and installs thin shell scripts for `post-merge`, `post-checkout`, and `post-rewrite`.

2. Each hook first tries the `autovendor` binary on your PATH. If not found, it falls back to `go run github.com/mpartipilo/autovendor@v{version}` — pinned to the exact version that was installed. This means hooks work even on machines without `autovendor` installed (only Go is required).

3. The hook calls `autovendor run <hook-type>`, which:
   - Determines the old and new git refs for the operation
   - Runs `git diff --name-only` to find changed `go.mod`/`go.sum` files
   - For each affected module directory that has a `vendor/` folder, runs `go mod vendor`

4. **Monorepo support:** If your repo has multiple Go modules (e.g., `services/auth/go.mod`, `tools/lint/go.mod`), only the modules whose dependencies changed are re-vendored.

5. **Plays nice with existing hooks:** If you already have git hooks, autovendor appends its block (wrapped in `# autovendor:begin/end` markers) without clobbering your existing scripts. Uninstall cleanly removes only the autovendor block.

## Example output

```text
autovendor: go.mod changed in ., running go mod vendor...
autovendor: vendor synced in . ✓
```

## Requirements

- Go (for `go mod vendor`)
- Git

## Security

The `go run` fallback in each hook is **pinned to the exact version** of autovendor that was used during `autovendor install`. For example, if you install with v1.0.0, the hooks will contain:

```sh
go run github.com/mpartipilo/autovendor@v1.0.0 run post-merge "$@"
```

This prevents supply chain attacks — a compromised future version cannot be silently pulled in by your hooks. To upgrade, run `autovendor uninstall && autovendor install` with the new version.

## FAQ

**Does it slow down git pull?**
No — it only runs `go mod vendor` when `go.mod` or `go.sum` actually changed. On pulls with no dependency changes, it's a no-op.

**What if I don't use vendoring?**
autovendor checks for a `vendor/` directory before running anything. If your module doesn't vendor, it's a no-op even if go.mod changed.

**Does it work with git worktrees?**
Yes — it uses `git rev-parse --git-dir` to find the correct hooks location.

**Can I use it alongside other git hook tools (lefthook, husky, pre-commit)?**
Yes — autovendor wraps its block in markers and appends to existing hooks rather than replacing them.

**How do I upgrade autovendor in an existing repo?**
Install the new version of autovendor, then re-install the hooks to update the pinned version:

```sh
autovendor uninstall
autovendor install
```

## License

MIT
