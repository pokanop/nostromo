---
title: Manifest YAML schema
---

# Manifest YAML schema

Manifests are plain YAML files stored under `~/.nostromo/ships/`. You rarely need to touch them — the CLI does it for you — but knowing the layout helps when hand-editing, reviewing a manifest before docking it, or authoring one to share.

!!! note "Key casing"
    `nostromo` writes every key in **lowercase** (`keypath`, `aliasonly`, `backupcount`) and expects the same casing when reading. `keyPath:` or `aliasOnly:` in a hand-edited file will be silently ignored.

## Full example

The manifest below was produced by:

```sh
nostromo add cmd k kubectl -d "kubectl shortcuts"
nostromo add cmd k.pods "get pods -n"
nostromo add sub k my-very-long-production-namespace prod
nostromo add cmd hello -l js -c 'console.log("hi")'
nostromo add cmd c clear -a
```

```yaml
name: manifest
source: file:///home/me/.nostromo/ships/manifest.yaml
path: /home/me/.nostromo/ships/manifest.yaml
version:
  uuid: dc90dc5f-9c35-437f-8763-f3967730b67f
  semver: 0.10.0
  gitcommit: 1a2b3c4
  builddate: 2024-01-01T00:00:00Z
config:
  verbose: false
  aliasesonly: false
  mode: 0
  backupcount: 10
commands:
  c:
    keypath: c
    name: clear
    alias: c
    aliasonly: true
    description: ""
    commands: {}
    subs: {}
    code:
      language: ""
      snippet: ""
    mode: 0
    disabled: false
  hello:
    keypath: hello
    name: ""
    alias: hello
    aliasonly: false
    description: ""
    commands: {}
    subs: {}
    code:
      language: js
      snippet: console.log("hi")
    mode: 0
    disabled: false
  k:
    keypath: k
    name: kubectl
    alias: k
    aliasonly: false
    description: kubectl shortcuts
    commands:
      pods:
        keypath: k.pods
        name: get pods -n
        alias: pods
        aliasonly: false
        description: ""
        commands: {}
        subs: {}
        code:
          language: ""
          snippet: ""
        mode: 0
        disabled: false
    subs:
      prod:
        name: my-very-long-production-namespace
        alias: prod
    code:
      language: ""
      snippet: ""
    mode: 0
    disabled: false
```

## Manifest

Top-level fields.

| Key        | Type                       | Description                                                                                                                                         |
| ---------- | -------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`     | string                     | Manifest name. `manifest` is reserved for the core manifest; docked manifests are saved as `ships/<name>.yaml` using this value                     |
| `source`   | string                     | Where the manifest came from — a `file://` URL for local manifests, otherwise the URL passed to `nostromo dock`. Used by `sync`                      |
| `path`     | string                     | Local path of the file. Rewritten by `nostromo` on save                                                                                              |
| `version`  | [Version](#version)        | Identity and build information                                                                                                                       |
| `config`   | [Config](#config)          | Settings. Only the **core** manifest's config is consulted for `get`/`set`, defaults and verbosity                                                   |
| `commands` | map of [Command](#command) | Top-level commands keyed by alias                                                                                                                    |

### Version

| Key         | Type   | Description                                                                                                                      |
| ----------- | ------ | -------------------------------------------------------------------------------------------------------------------------------- |
| `uuid`      | string | Unique identifier regenerated whenever the manifest changes. `sync` only replaces a manifest when this differs from the local copy |
| `semver`    | string | `nostromo` version that last wrote the file                                                                                        |
| `gitcommit` | string | Git commit of that build                                                                                                           |
| `builddate` | string | Build date of that build                                                                                                           |

### Config

| Key           | Type    | Default | Description                                                                          |
| ------------- | ------- | ------- | ------------------------------------------------------------------------------------ |
| `verbose`     | bool    | `false` | Always log verbosely                                                                 |
| `aliasesonly` | bool    | `false` | New commands are [alias-only](concepts/alias-only.md) by default                     |
| `mode`        | int     | `0`     | Default [execution mode](concepts/modes.md): `0` concatenate, `1` independent, `2` exclusive |
| `backupcount` | int     | `10`    | Number of [backups](concepts/backups.md) kept per manifest; `0` disables backups     |

See [Configuration](concepts/configuration.md) for the CLI side of these.

### Command

Commands nest recursively through `commands`. The map key is always the command's `alias`.

| Key           | Type                              | Description                                                                                                                                 |
| ------------- | --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `keypath`     | string                            | Dot-delimited path from the root, e.g. `k.pods`. Maintained by `nostromo`; must match the command's position in the tree                     |
| `name`        | string                            | The shell command fragment to run, e.g. `get pods -n`. May be empty when `code` is set                                                       |
| `alias`       | string                            | The word you type. For root commands this becomes a shell function (or alias)                                                                |
| `aliasonly`   | bool                              | Create a plain shell alias instead of a `nostromo` function. See [Alias-only commands](concepts/alias-only.md)                              |
| `description` | string                            | Human readable description shown by `show` and in completions                                                                                |
| `commands`    | map of Command                    | Sub-commands keyed by alias                                                                                                                  |
| `subs`        | map of [Substitution](#substitution) | Substitutions available at this node and below, keyed by alias                                                                            |
| `code`        | [Code](#code)                     | Optional code snippet that replaces `name`                                                                                                   |
| `mode`        | int                               | Execution mode for this node: `0` concatenate, `1` independent, `2` exclusive                                                                |
| `disabled`    | bool                              | When `true`, this command and everything beneath it refuse to run (`command is disabled at <keypath>`). Not settable from the CLI; edit by hand |

### Substitution

| Key     | Type   | Description                                              |
| ------- | ------ | -------------------------------------------------------- |
| `name`  | string | The original, long value that is inserted into the command |
| `alias` | string | The short token you type in its place                     |

### Code

| Key        | Type   | Description                                                |
| ---------- | ------ | ---------------------------------------------------------- |
| `language` | string | One of `sh`, `ruby`, `python`, `perl`, `js`                  |
| `snippet`  | string | The code to run. Use a YAML block scalar (`|`) for multiple lines |

A code block is only used when **both** `language` and `snippet` are non-empty.

## Spaceport file

`~/.nostromo/spaceport.yaml` tracks which manifests are docked and in what order, plus the theme:

```yaml
sequence:
- manifest
- edit
- docker
theme: 2
```

| Key        | Type            | Description                                                                                 |
| ---------- | --------------- | ------------------------------------------------------------------------------------------- |
| `sequence` | list of strings | Manifest names in lookup order. Commands are resolved from the first manifest that has them |
| `theme`    | int             | `0` default, `1` grayscale, `2` emoji                                                        |

## Validation

A file is only accepted as a manifest if it parses as YAML and has a non-empty `name`. Missing `config` values fall back to the defaults above; `path` is always replaced with the file's actual location on load. Anything docked must satisfy the same rules, so `nostromo dock` on a folder simply skips files that aren't manifests.

## Migration from older layouts

`nostromo init` moves a core manifest found at the pre-spaceport location `~/.nostromo/manifest.yaml` into `ships/` and renames an old `backups/` folder to `cargo/`. Manifests written by older versions with a `file:/path` source are normalized to `file:///path` on load.
