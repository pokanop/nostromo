---
title: Backups (cargo)
---

# Backups (cargo)

Backups of manifests are taken automatically to prevent data loss in case of shenanigans gone wrong. Before a manifest file is overwritten — by `add`, `remove`, `update`, `move`, `dock`, `sync`, ... — the previous version is copied to the **cargo** folder:

```text
~/.nostromo/cargo/
├── manifest_1789013002053.yaml
├── manifest_1789013002063.yaml
├── edit_1789013002074.yaml
└── ...
```

Files are named `<manifest name>_<unix milliseconds>.yaml`, so they sort chronologically and each manifest's backups are easy to tell apart.

## Retention

The maximum number of backups **per manifest** is controlled by the `backupCount` setting, which each manifest carries in its `config` block (the core manifest's value is the one you set from the CLI). The default is `10`.

```sh
nostromo set backupCount 20
nostromo get backupCount
```

When a new backup is written, the oldest ones beyond the limit are pruned. Setting `backupCount` to `0` disables backups entirely (and removes existing ones for that manifest on the next save).

## Restoring

Backups are ordinary manifests, so restoring is a copy:

```sh
cp ~/.nostromo/cargo/manifest_1789013002053.yaml ~/.nostromo/ships/manifest.yaml
nostromo init
```

Run `nostromo init` afterwards to regenerate completions. If you restore a docked manifest that you also `sync`, remember that the next sync may overwrite it again if the upstream identifier differs.

!!! note "Shell startup files"
    Your shell startup file gets its own safety net: before `nostromo` rewrites `.bashrc`, `.zshrc`, `config.fish` or your PowerShell profile it saves a timestamped copy (e.g. `.zshrc_20240101120000`) to your system temp directory.

## Manifest migrations

`nostromo init` also migrates older layouts: a core manifest found at `~/.nostromo/manifest.yaml` is moved into `ships/`, and an old `backups/` folder is renamed to `cargo/`.
