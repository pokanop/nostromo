---
title: Getting started
---

# Getting started

Getting `nostromo` up and running takes three steps:

1. [Install](installation.md) the `nostromo` binary.
2. Run `nostromo init` and add your first command — see the [Quick start](quickstart.md).
3. Make sure your shell picks up the generated functions and completions — see [Shell integration](shell-integration.md).

## Prerequisites

- Works for macOS and Linux with `bash`, `zsh`, `fish` and PowerShell (`pwsh`) shells.
- Windows with PowerShell should work *but is untested*.

## How it works

`nostromo` is not a shell builtin, so it uses a small amount of glue to make its magic work:

- Your commands live in YAML **manifests** under `~/.nostromo/ships/`.
- `nostromo init` adds a `# nostromo [section begin]` … `# nostromo [section end]` block to your shell startup file that sources `nostromo completion <shell>`.
- That completion script defines a shell function for each top level command. When you run `foo bar baz`, the function calls `nostromo eval foo bar baz`, which resolves the command tree, applies substitutions and prints the final shell command. The function then runs that command with `eval` in your current shell.
- Because commands are evaluated in your shell (not a subprocess), things like `cd` and `export` persist — just like a regular alias.
- Completions are reloaded automatically after every `nostromo` command, so new commands are available immediately.

Read on to [install](installation.md) `nostromo`.
