---
title: copy cmd
---

# nostromo copy cmd

Copy a command in nostromo manifest

## Synopsis

Copy a command in nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

Manipulate nostromo manifests by copying command nodes. Using copy, take
a source key path and copy it to the destination key path with all 
subcommands in tow.

If the destination key path does not exist, it will be created
and the node will be copied there. Optionally provide a description
with -d to the destination.

To copy a node to the root, use '.' for destination. Optionally provide
a destination manifest using the -m flag. Otherwise the copy will default 
to the core manifest.

```
nostromo copy cmd [src.key.path] [dest.key.path] [options] [flags]
```

## Options

```
  -d, --description string   Description of the destination to copy command to
  -h, --help                 help for cmd
  -m, --manifest string      Destination manifest to copy command to
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo copy](nostromo_copy.md)	 - Copy a command in nostromo manifest

