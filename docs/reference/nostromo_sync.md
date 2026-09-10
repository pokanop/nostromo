---
title: sync
---

# nostromo sync

Sync docked manifests from source locations

## Synopsis

Sync docked manifests from source locations and
makes commands available for execution.

Sync can be used to update previously docked manifests from
respective data sources. Provide one or more of the names of 
the manifests as arguments to sync.

Sync will only update manifests with changed identifiers, to
force update use the -f flag.

```
nostromo sync [name]... [flags]
```

## Options

```
  -f, --force   Force sync manifests
  -h, --help    help for sync
  -k, --keep    Keep downloaded files
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

