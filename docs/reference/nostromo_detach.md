---
title: detach
---

# nostromo detach

Detach a command node into a new manifest

## Synopsis

Detach will extract an entire command tree from the given node and
create a brand new manifest under the ships/ folder.

Running:

	nostromo detach mobile-builds build.ios build.android

creates a new mobile-builds.yaml file with build command sets joined in the
new manifest. If a file with the target name already exists, nostromo will
attempt to merge the manifests automatically.

By default, the detached node is removed from the manifest. If this node should
be kept in the original manifest, use the --keep flag.

This is a convenient way to slice your command sets and produce manifests
that can be shared via the sync command.

```
nostromo detach [name] [key.path]... [options] [flags]
```

## Options

```
  -d, --description string   A description for the destination key path
  -h, --help                 help for detach
  -k, --keep                 Keep original command tree intact
  -r, --root string          Add detached commands to a key path, defaults to root
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

