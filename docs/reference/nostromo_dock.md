---
title: dock
---

# nostromo dock

Dock nostromo manifests

## Synopsis

Dock nostromo manifests from source locations and make commands
available for execution.


Dock can be used to copy a single manifest or more to nostromo's config
folder. If a docked manifest already exists with the same name then
nostromo will overwrite that file if the identifier is different.

Run:

	nostromo dock http://foo.com/edit.yaml file://path/to/install.yaml

To force docking even if identifiers are the same, use the -f flag.

```
nostromo dock [source]... [options] [flags]
```

## Options

```
  -f, --force   Force dock manifest
  -h, --help    help for dock
  -k, --keep    Keep downloaded files
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

