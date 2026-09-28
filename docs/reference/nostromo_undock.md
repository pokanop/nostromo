---
title: undock
---

# nostromo undock

Undock nostromo manifests

## Synopsis

Undock nostromo manifests and remove commands from being
executable.

A manifest added to nostromo using the dock command can be undocked. 
This will delete the file from the local configuration and the commands
will no longer be available to run. Manifests only pulled in through the
links of the undocked manifest are removed as well, while manifests still
linked from another manifest are kept.

Run:

	nostromo undock edit install

To get the commands back you will need to dock the manifest again from 
the original source location.

```
nostromo undock [name]... [flags]
```

## Options

```
  -h, --help   help for undock
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

