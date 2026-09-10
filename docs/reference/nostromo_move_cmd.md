---
title: move cmd
---

# nostromo move cmd

Move a command in nostromo manifest

## Synopsis

Move a command in nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

Manipulate nostromo manifests by moving command nodes. Using move, take
a source key path and shift it to the destination key path with all 
subcommands in tow.

If the destination key path does not exist, it will be created
and the node will be moved there. Optionally provide a description
with -d to the destination.

To move a node to the root, use '.' for destination. Optionally provide
a destination manifest using the -m flag. Otherwise the move will default 
to the core manifest.

```
nostromo move cmd [src.key.path] [dest.key.path] [options] [flags]
```

## Options

```
  -d, --description string   Description of the destination to move command to
  -h, --help                 help for cmd
  -m, --manifest string      Destination manifest to move command to
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo move](nostromo_move.md)	 - Move a command in nostromo manifest

