---
title: rename cmd
---

# nostromo rename cmd

Rename a command in nostromo manifest

## Synopsis

Rename a command in nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

Manipulate the core nostromo manifest by renaming command nodes.
Using rename, take a source key path and rename it to a new name with
all subcommands in tow.

If the new name already exists, the command will fail.

Optionally provide a description with -d to the destination.

```
nostromo rename cmd [key.path] [name] [options] [flags]
```

## Options

```
  -d, --description string   Description of the command to rename
  -h, --help                 help for cmd
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo rename](nostromo_rename.md)	 - Rename a command in nostromo manifest

