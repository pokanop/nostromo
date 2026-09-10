---
title: Quick start
---

# Quick start

## Initialize

This command initializes `nostromo` and creates the core manifest under `~/.nostromo`:

```sh
nostromo init
```

`init` does a few things:

- creates the `~/.nostromo` folder with `ships/manifest.yaml` (the core manifest) and `spaceport.yaml`;
- writes completion scripts for bash, zsh, fish and PowerShell to `~/.nostromo/completions/` and man pages to `~/.nostromo/man/`;
- adds a `nostromo` block to your shell startup file (see [Shell integration](shell-integration.md)).

To customize the directory, set the `NOSTROMO_HOME` environment variable to a location of your choosing before running any `nostromo` command:

```sh
export NOSTROMO_HOME=~/.config/nostromo
nostromo init
```

!!! tip "Run `init` after every update"
    With every update it's a good idea to run `nostromo init` to ensure any manifest changes are migrated and commands continue to work. `nostromo` performs any migrations to files and folders at this time.

Then either restart your shell or source your startup file, e.g. `source ~/.zshrc`.

## Add your first command

```sh
nostromo add cmd foo "echo bar"
```

And just like that you can now run `foo` like any other alias:

```sh
$ foo
bar
```

Descriptions show up in tab completion, so add them where you can:

```sh
nostromo add cmd foo "echo bar" -d "My magical foo command that prints bar"
```

## Build a command tree

Use a dot-delimited **keypath** to nest commands:

```sh
nostromo add cmd foo.bar.baz "echo hello"
```

This builds the tree `foo` :material-arrow-right: `bar` :material-arrow-right: `baz`, and all of these are now valid (the first two do nothing *yet*):

```sh
foo
foo bar
foo bar baz
```

By default the commands along the path are concatenated, so given:

```sh
nostromo add cmd foo 'echo oof'
nostromo add cmd foo.bar 'rab'
nostromo add cmd foo.bar.baz 'zab'
```

running `foo bar baz` executes `echo oof rab zab`. See [Commands and keypaths](../concepts/commands.md) and [Execution modes](../concepts/modes.md) for the details.

## Interactive mode

Run `nostromo add` with no arguments to be walked through adding a command or substitution with prompts:

```sh
nostromo add
```

<p align="center">
  <img src="https://raw.githubusercontent.com/pokanop/nostromo/main/images/interactive.gif" alt="interactive" style="border-radius: 15px">
</p>

## Have a look around

```sh
nostromo show          # summary of manifests, config, commands and your profile block
nostromo show --tree   # command tree
nostromo show --yaml   # raw manifest yaml
nostromo show --json   # raw manifest json
nostromo find baz      # search commands and substitutions by name
nostromo eval foo bar  # print what would run, without running it
```

## Dock a ready-made manifest

The quickest way to populate your commands is the `dock` feature, which downloads a manifest from a local or remote source and makes its commands available:

```sh
nostromo dock https://github.com/pokanop/nostromo/raw/main/examples/edit.yaml
```

See [Examples](../examples.md) for the manifests shipped with the repo and [Dock, undock and sync](../concepts/distributed-manifests.md) for the supported sources.

## Start over

To destroy the core manifest and start fresh:

```sh
nostromo destroy
```

Add `-n`/`--nuke` to delete the entire `~/.nostromo` installation. Note that this does not remove the block `nostromo` added to your shell startup file; delete that by hand if you want it gone. Backups of your manifests are kept in `~/.nostromo/cargo` — see [Backups](../concepts/backups.md).
