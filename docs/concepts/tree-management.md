---
title: Tree management
---

# Tree management

Moving and copying command subtrees can be done with `nostromo` to avoid manual copy-pasta in YAML. All of these commands take [keypaths](commands.md#keypaths) and carry every sub-command and substitution beneath the node along for the ride.

## Move

```sh
nostromo move cmd <src.key.path> <dest.key.path>
```

The node at the source keypath is re-parented under the destination. If the destination keypath doesn't exist it's created; give the new destination a description with `-d`. Use `.` as the destination to move a node to the root:

```sh
nostromo move cmd tools.docker .        # tools.docker becomes docker
nostromo move cmd docker infra -d "Infrastructure tools"
```

Moves default to the core manifest; use `-m <manifest>` to move a node into a different docked manifest.

## Copy

```sh
nostromo copy cmd <src.key.path> <dest.key.path>
```

Same as move, but the source node stays where it is. `copy` also accepts `-d` and `-m`, so it's the easy way to replicate a branch into another manifest:

```sh
nostromo copy cmd build.ios . -m mobile
```

## Rename

```sh
nostromo rename cmd <key.path> <name>
```

Renames the node (its alias) in place, keeping its children and substitutions. If a sibling with the new name already exists the command fails. Add `-d` to change the description at the same time.

```sh
nostromo rename cmd foo.bar baz -d "Now called baz"
```

## Detach

So you've created an awesome suite of commands and you'd like to share, am I right? `detach` slices one or more command nodes out of your manifests into a brand new manifest under `~/.nostromo/ships/`:

```sh
nostromo detach <name> <key.path>...
```

For example:

```sh
nostromo detach mobile-builds build.ios build.android
```

creates `~/.nostromo/ships/mobile-builds.yaml` containing both command trees at its root, and docks it in the spaceport. If a manifest with that name already exists, the commands are merged into it.

| Flag                 | Purpose                                                                            |
| -------------------- | ---------------------------------------------------------------------------------- |
| `-k, --keep`         | Keep the original nodes in place (by default they're removed from the source)      |
| `-r, --root <key.path>` | Attach the detached commands under this keypath in the new manifest instead of its root |
| `-d, --description`  | Description for the `--root` keypath                                                |

!!! warning
    Detaching nodes from a *docked* manifest may have unwanted side effects: the next `nostromo sync` will most likely add them back from the original source.

The resulting file is a self-contained manifest you can publish for others to [dock](distributed-manifests.md#dock).

## Regenerate an identifier

`nostromo` updates a manifest's identifier every time a command changes it. If you edit a YAML file by hand you can bump the identifier yourself so that `sync` on other machines picks up the change:

```sh
nostromo uuidgen <name>   # omit the name for the core manifest
```

## Remove

Remove a node and everything beneath it, or a single substitution:

```sh
nostromo remove cmd foo.bar
nostromo remove sub foo.bar sls
```
