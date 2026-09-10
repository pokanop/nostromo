---
title: Manifests and the spaceport
---

# Manifests and the spaceport

## Manifests (ships)

A **manifest** is a YAML file that holds a tree of [commands](commands.md) and a handful of [config](configuration.md) settings. Manifests live in the `ships/` folder of your `nostromo` home:

```text
~/.nostromo/
├── ships/
│   ├── manifest.yaml     # the core manifest
│   ├── edit.yaml         # docked manifests
│   └── docker.yaml
├── spaceport.yaml        # load order and theme
├── cargo/                # backups
├── completions/          # generated completion scripts
└── man/                  # generated man pages
```

The **core manifest** is `manifest.yaml`. It's created by `nostromo init` and is where `add`, `remove`, `update` and friends write by default. Every other file in `ships/` is a **docked** manifest (see [Dock, undock and sync](distributed-manifests.md)). You can add as many manifests as you like and `nostromo` parses and aggregates all of their commands — handy for organizations that want to build their own command suite.

Set `NOSTROMO_HOME` to move the whole folder somewhere else.

A manifest looks like this (see the [schema](../schema.md) for every field):

```yaml
name: manifest
source: file:///Users/gopher/.nostromo/ships/manifest.yaml
path: /Users/gopher/.nostromo/ships/manifest.yaml
version:
  uuid: 4b897259-290b-4975-a7a9-a37e6813273d
  semver: 0.9.9
  gitcommit: 2e19036fcc51c9af254ff0fa5fa849558308a985
  builddate: '2022-02-06T08:21:54Z'
config:
  verbose: false
  aliasesonly: false
  mode: 0
  backupcount: 10
commands:
  foo:
    keypath: foo
    name: echo bar
    alias: foo
    aliasonly: false
    description: My magical foo command
    commands: {}
    subs: {}
    code:
      language: ''
      snippet: ''
    mode: 0
    disabled: false
```

### Versions and identifiers

Each manifest carries a `version` block. The `semver`, `gitcommit` and `builddate` come from the `nostromo` binary that last wrote the file. The `uuid` is the manifest's **identifier**: every `nostromo` command that modifies a manifest regenerates it, and [`sync`](distributed-manifests.md#sync) uses it to decide whether a docked manifest has changed upstream. If you edit a manifest by hand and want to publish it, run `nostromo uuidgen <name>` to bump the identifier.

## The spaceport

The **spaceport** is where manifests dock. It's stored in `~/.nostromo/spaceport.yaml`:

```yaml
sequence:
- manifest
- edit
- docker
theme: 2
```

- `sequence` is the order manifests are loaded in. When a command is looked up, manifests are searched in this order and the first match wins, so the core manifest always takes precedence over docked ones.
- `theme` is the active [theme](configuration.md#themes).

Docking and undocking keep the sequence up to date; you shouldn't need to edit this file by hand. If it's missing, `nostromo` recreates it from the files in `ships/`.

## Inspecting manifests

```sh
nostromo show           # every manifest with its config, commands and your profile block
nostromo show --tree    # command tree across all manifests
nostromo show --yaml    # raw yaml
nostromo show --json    # raw json
```

<p align="center">
  <img src="https://raw.githubusercontent.com/pokanop/nostromo/main/images/tree.gif" alt="tree" style="border-radius: 15px">
</p>

Turn on the `verbose` config setting (or pass `-v`) for more detail on every command:

<p align="center">
  <img src="https://raw.githubusercontent.com/pokanop/nostromo/main/images/verbose.gif" alt="verbose" style="border-radius: 15px">
</p>

## Editing manifests by hand

Manifests are plain YAML, so you can edit them directly — for example to write a multi-line [code snippet](code-snippets.md) or to `disable` a command. Keep these in mind:

- Keys are **lowercase** (`aliasonly`, `backupcount`), as written by `nostromo`.
- Every command needs its `keypath` to match its position in the tree; `nostromo` uses it for lookups.
- After hand edits, run `nostromo init` (or any `nostromo` command) so completions are regenerated, and `nostromo uuidgen <name>` if you plan to share the manifest.
- A backup is taken before every write — see [Backups](backups.md).
