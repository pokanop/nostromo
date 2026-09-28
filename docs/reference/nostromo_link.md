---
title: link
---

# nostromo link

Link a manifest as a dependency

## Synopsis

Link a manifest from a source location as a dependency of another
manifest, docking it along with any manifests it links itself.

The link is recorded on the target manifest, which is the core manifest
unless --to is given, so syncing the target fetches the linked manifest
as well. Linking fails if a manifest with the same name already exists
from a different source or if the link would be circular.

Run:

	nostromo link https://foo.com/edit.yaml
	nostromo link file://path/to/install.yaml --to tools

Use unlink to remove the dependency again.

```
nostromo link [source] [options] [flags]
```

## Options

```
  -f, --force       Force update of the linked manifest
  -h, --help        help for link
  -k, --keep        Keep downloaded files
  -t, --to string   Manifest to record the link on (default: core manifest)
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

