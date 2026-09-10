---
title: Alias-only commands
---

# Alias-only commands

By default every `nostromo` command is a shell **function** that calls back into the `nostromo` binary: the function for the root command `foo` passes the remaining arguments to `nostromo eval`, which resolves the command tree and prints a command that the function then runs with `eval`. That's what makes keypaths, substitutions and modes work.

Sometimes you just want a boring, standard shell alias. `nostromo` can manage those for you too.

## Creating an alias-only command

Either pass the `--alias-only` / `-a` flag when adding a command:

```sh
nostromo add cmd c "clear" --alias-only
```

or set the `aliasesOnly` config setting so that *every* command added afterwards is a plain alias:

```sh
nostromo set aliasesOnly true
nostromo add cmd c "clear"
```

Both produce this line in the sourced shell script:

```sh
alias c='clear'
```

instead of the function a regular `nostromo` command gets:

```sh
c() { eval $(__nostromo_cmd eval c "$@"); }
```

In fish the alias is the same; in PowerShell, where `Set-Alias` can't carry arguments, it becomes `function c { clear @args }`.

## What you give up

- **No command tree.** Keypaths have no effect on an alias-only command — it's always a single root level alias. `nostromo add cmd foo.bar.baz "cd /tmp" -a` literally creates `alias foo.bar.baz='cd /tmp'`.
- **No substitutions, modes or code snippets.** The shell runs the alias text as-is; `nostromo` isn't involved at execution time.
- **Completion** for the alias's arguments is whatever your shell provides for the underlying command.

## What you keep

- The alias is stored in your manifest, so it's backed up, shows up in `nostromo show`, can be docked, synced, and shared like any other command.
- `nostromo update`, `rename`, `remove` and friends work on it.

!!! tip
    Alias-only commands are a good fit for one-word shortcuts (`c` for `clear`, `g` for `git`). Use regular `nostromo` commands whenever you want to compose things.

In the manifest, an alias-only command has `aliasonly: true`:

```yaml
commands:
  c:
    keypath: c
    name: clear
    alias: c
    aliasonly: true
    description: Clear the console
```
