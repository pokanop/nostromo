---
title: unlink
---

# nostromo unlink

Unlink a manifest dependency

## Synopsis

Unlink a manifest dependency from another manifest.

The link is removed from the target manifest, which is the core manifest
unless --to is given. The linked manifest is undocked when nothing else
links it and it was not docked directly, along with any manifests only
it pulled in.

Run:

	nostromo unlink edit
	nostromo unlink install --to tools

```
nostromo unlink [name] [options] [flags]
```

## Options

```
  -h, --help        help for unlink
  -t, --to string   Manifest to remove the link from (default: core manifest)
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

