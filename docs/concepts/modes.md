---
title: Execution modes
---

# Execution modes

A command's **mode** controls how it combines with the other commands along its keypath when `nostromo` builds the final command line. There are three modes:

| Mode          | Behaviour                                                                                     |
| ------------- | --------------------------------------------------------------------------------------------- |
| `concatenate` | **Default.** Join this command with its parents and children, separated by spaces.            |
| `independent` | Terminate this command with `;` so it runs on its own before the rest of the keypath.         |
| `exclusive`   | Run this command and only this command, ignoring every parent command.                        |

## Concatenate

By default `nostromo` walks the keypath from the root down and concatenates each node's command. Given:

```sh
nostromo add cmd foo 'echo oof'
nostromo add cmd foo.bar 'rab'
nostromo add cmd foo.bar.baz 'zab'
```

running `foo bar baz` executes:

```sh
echo oof rab zab
```

This is what lets you build up a single long invocation piece by piece — think `docker` :material-arrow-right: `docker compose` :material-arrow-right: `docker compose up`.

## Independent

Marking a node `independent` appends a `;` after its command, so it executes as a separate statement before whatever comes next down the keypath:

```sh
nostromo add cmd foo 'echo oof'
nostromo add cmd foo.bar 'cd /tmp' -m independent
nostromo add cmd foo.bar.baz 'ls'
```

`foo bar baz` now runs:

```sh
echo oof cd /tmp; ls
```

Typically you'd make the root independent too so each step stands alone. A common pattern is a root command that `cd`s into a repo and children that run tools inside it:

```sh
nostromo add cmd build 'cd ~/src/app' -m independent
nostromo add cmd build.ios 'xcodebuild -scheme app'
nostromo add cmd build.android './gradlew build'
```

## Exclusive

An `exclusive` command ignores its parents entirely; only its own command runs:

```sh
nostromo add cmd edit 'vim' 
nostromo add cmd edit.ssh 'vim ~/.ssh/config'
nostromo add cmd edit.ssh.hosts 'vim ~/.ssh/hosts' -m exclusive
```

`edit ssh hosts` runs just `vim ~/.ssh/hosts` — not `vim vim ~/.ssh/config vim ~/.ssh/hosts`. This is useful for grouping related commands under a namespace whose parents shouldn't contribute anything.

!!! note "Exclusive only applies to the command you run"
    The `exclusive` check is made on the *last* node of the keypath you invoke. If an `exclusive` node has children and you run one of them, the exclusive node is treated like `independent` (its command is terminated with `;`) so that the child still runs.

## Setting the mode

Per command, when adding or updating:

```sh
nostromo add cmd foo.bar.baz -m exclusive "echo baz"
nostromo update foo.bar.baz -m independent
```

Or globally, as the default for every command added afterwards:

```sh
nostromo set mode independent
```

The global setting is stored in the core manifest's `config.mode` and is applied whenever a command is added or updated without an explicit `-m`; existing commands are otherwise left alone. Check the mode of any command with `nostromo show --yaml` — it's stored as a number: `0` concatenate, `1` independent, `2` exclusive.

Use `nostromo eval <keypath...>` to preview the exact command line any combination of modes produces.
