---
title: links
---

# nostromo links

Show the manifest dependency graph

## Synopsis

Show the dependency graph of linked manifests.

Without a name every docked manifest is shown along with the manifests
it links. Pass a manifest name to show only its dependencies. Circular
references and links whose manifest is not docked are marked.

Run:

	nostromo links
	nostromo links tools

```
nostromo links [name] [flags]
```

## Options

```
  -h, --help   help for links
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

