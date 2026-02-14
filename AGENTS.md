# hackctl

A CLI tool for managing developer workspaces. Automates git worktrees, tmux sessions, and Claude Code environments so you can go from "I want to work on X" to a fully bootstrapped workspace in one command.

## How It Works

A **workspace** is a directory on disk under `~/.hackctl/`. The filesystem is the sole source of truth — if the directory exists, the workspace exists. There is no database.

There are two types:

- **Worktree**: A git worktree inside a bare clone. For writing code.
- **Temporary**: A plain directory. For non-code work like Claude discussions or MCP sessions. Auto-cleaned after a configurable inactivity period (default: 2 weeks).

A **tmux session** is derived from a workspace on demand — it is a view, not part of the workspace's identity. Opening a workspace creates or attaches to its tmux session transparently.

## Disk Layout

```
~/.hackctl/
├── config.yaml
├── repositories/
│   └── <org>/
│       └── <repo>/              # bare clone
│           └── <branch>/        # worktree
└── temporary/
    └── <name>/                  # temporary workspace
```

`~/.hackctl/` and `config.yaml` are created implicitly on first use. No `init` command.

## Workspace Naming

Every workspace has a globally unique name, derived from its filesystem path. The name is never parsed back into components — hackctl discovers workspaces by walking the filesystem and computing names from the directory structure.

| Type | Path | Name |
|---|---|---|
| Worktree | `repositories/<org>/<repo>/<branch>/` | `<org>-<repo>-<branch>` |
| Temporary | `temporary/<name>/` | `<name>` |

When no branch or name is provided at creation time, one is auto-generated using a D&D monster name with an adjective prefix (e.g., `swift-owlbear`, `fuzzy-beholder`). For worktree workspaces, this also becomes the git branch name.

The workspace name is used:
- As the CLI argument to identify a workspace
- To derive the tmux session name: `hackctl-<name>` (non-alphanumeric characters replaced with hyphens)
- In `list` output (fzf-friendly, one workspace per line)

## CLI Design

Verb-noun pattern using cobra (e.g., `create workspace`, `list workspaces`, `open workspace`). A `--type` flag distinguishes workspace types at creation time. The code is the reference for exact command syntax.

### Behaviours

- **Creating (worktree)**: Bare-clone the repo if needed. Create the branch from the default branch if it doesn't exist on the remote. Create the worktree.
- **Creating (temporary)**: Create the directory.
- **Listing**: Discover all workspaces from the filesystem. Show inline status: merge status, tmux session state, and potentially CI status or whether Claude needs input.
- **Opening**: Idempotent. Attach to existing tmux session, or create one from the layout config and attach.
- **Deleting**: Warn and require confirmation if there are uncommitted changes or unpushed commits (or accept a force flag). Clean up worktree, tmux session, and directory. If this was the last worktree in a bare clone, delete the bare clone too (after verifying no other branches have unpushed work).
- **Renaming**: Rename on disk (and the git branch, if applicable). The workspace name updates automatically since it's derived from the filesystem.

### Error handling

Fail fast and loud. Print a clear error, clean up partial state, exit non-zero.

## Tmux

### Layouts

Defined in `config.yaml` per workspace type. A layout is a list of windows. Each window has a name and one or two commands. Two commands splits the window horizontally.

```yaml
layouts:
  worktree:
    - name: claude
      command: claude
    - name: diff
      command: watch -n 5 git diff
  temporary:
    - name: claude
      command: claude
```

### Quick switcher

hackctl integrates with tmux popup and fzf — it does not ship its own TUI. A tmux keybinding invokes a popup that lists workspaces via fzf, then opens the selection. The popup should also support creating new workspaces.

### Cleanup

Orphaned tmux sessions (no corresponding workspace on disk) are cleaned up automatically on every hackctl invocation.

## Architecture

- **Language**: Go
- **CLI framework**: cobra
- **Git**: Shell out to `git`. User's git config handles authentication.
- **Tmux**: Shell out to `tmux`.

## Conventions

- `gofmt` enforced, `golangci-lint` with strict rules.
- TDD. Write tests first. Tests in `_test.go` files alongside the code they test.
- Return errors, don't panic. Wrap with context: `fmt.Errorf("doing thing: %w", err)`.
- Standard Go naming. No stutter (`workspace.New()` not `workspace.NewWorkspace()`).
- Keep it flat until complexity demands packages.
- Prefer shelling out to git/tmux over libraries.
- Plain text output, designed for piping and composability with fzf.

## Build

```
make build       # go build
make test        # go test ./...
make lint        # golangci-lint run
make fmt         # gofmt -w .
```

## Philosophy

- Solo project. Optimise for the author's workflow, not generalisation.
- Don't over-abstract early. Start concrete, extract when patterns repeat.
- Fast. No unnecessary network calls.
- Workspaces are cheap and disposable.
