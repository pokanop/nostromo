---
title: uuidgen
---

# nostromo uuidgen

Generate a new unique id for a manifest

## Synopsis

Generate a new unique id for a manifest.

nostromo uses a uuid to determine if a manifest is unique or not.
When using sync to get new manifests, nostromo will only apply the
changes if it detects a different identifier.

This command allows regenerating the uuid to allow for publishing
and pulling updated manifests. Note that using standard nostromo
commands automatically updates the identifier. This command can be
used if manual updates were made to a manifest.

Omitting the name of the manifest will apply to the core manifest.

```
nostromo uuidgen [name] [flags]
```

## Options

```
  -h, --help   help for uuidgen
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

