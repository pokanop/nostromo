---
title: Dock, undock and sync
---

# Dock, undock and sync

`nostromo` supports keeping multiple manifest sources :muscle:, letting you organize and distribute your commands as you please. Manifests can be fetched from many data sources:

- Local files
- Git
- Mercurial
- HTTP
- Amazon S3
- Google GCS

!!! info
    `nostromo` uses [go-getter](https://github.com/hashicorp/go-getter) for downloading. Details on supported URL formats, protocols and options (branches, subdirectories, checksums, ...) are in the go-getter documentation.

## Dock

To add — **dock** — one or more manifests:

```sh
nostromo dock <source>...
```

For example:

```sh
nostromo dock https://github.com/pokanop/nostromo/raw/main/examples/edit.yaml
nostromo dock file://path/to/install.yaml
nostromo dock github.com/pokanop/nostromo//examples
```

And that's it! Your shell now has the new manifest's commands.

What happens under the hood:

1. Each source is downloaded into a temporary folder under `~/.nostromo/downloads`. Regular `github.com/.../blob/...` links are rewritten to their `raw` equivalent for you.
2. Every YAML file found in the download is parsed as a manifest. A source can therefore be a single file *or* a folder / repository containing several manifests.
3. Each manifest is saved to `~/.nostromo/ships/<name>.yaml`, where `<name>` is the `name` field **inside** the file, and it's added to the [spaceport](manifests.md#the-spaceport). A downloaded manifest named `manifest` (the core manifest's name) is renamed to `<filename>-<timestamp>` so it can't clobber yours.
4. The manifest's `source` is set to the URL you docked, so it can be synced later.
5. The download folder is deleted unless you pass `-k` / `--keep`.

If a docked manifest with the same name already exists, `nostromo` only overwrites it when the version identifier is different. Use `-f` / `--force` to overwrite regardless:

```sh
nostromo dock -f <source>
```

| Flag           | Purpose                                           |
| -------------- | ------------------------------------------------- |
| `-f, --force`  | Dock even if the identifier hasn't changed        |
| `-k, --keep`   | Keep the downloaded files in `~/.nostromo/downloads` |

## Sync

To update previously docked manifests from their sources, run `sync` with the names of the manifests. Omit the names to sync **all** docked manifests:

```sh
nostromo sync              # everything
nostromo sync edit docker  # just these
```

`nostromo` syncs manifests using the `version.uuid` identifier in the manifest and only rewrites a manifest when the identifier differs from the local copy. To force an update anyway:

```sh
nostromo sync -f <name>...
```

`sync` takes the same `-f` / `-k` flags as `dock`. The core manifest is never synced.

!!! tip "Publishing updates"
    Every `nostromo` command that changes a manifest regenerates its identifier automatically. If you edit a shared manifest **by hand**, run `nostromo uuidgen <name>` before pushing it so that consumers' `nostromo sync` picks up the change.

## Undock

If you're tired of someone else's manifest or it just isn't making you happy :frowning: then undock it:

```sh
nostromo undock <name>...
```

This deletes the file from `~/.nostromo/ships`, removes it from the spaceport, and its commands disappear from your shell. To get them back you'll need to dock the manifest again from its original source.

## Sharing your own commands

Slice a subtree of your core manifest into a new manifest with [`detach`](tree-management.md#detach) and publish the resulting `~/.nostromo/ships/<name>.yaml` anywhere go-getter can reach — a Git repository is the natural choice. Others then run:

```sh
nostromo dock github.com/you/your-manifests//path/to/name.yaml
```

Manifests are searched in spaceport order, so if a docked manifest defines a command with the same name as one of yours, yours (in the core manifest) wins.
