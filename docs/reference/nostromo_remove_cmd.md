---
title: remove cmd
---

# nostromo remove cmd

Remove a command from nostromo manifest

## Synopsis

Remove a command to nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.
	
This will remove appropriate command scopes for all levels beneath
the provided key path. A command scope can a tree of sub commands
and substitutions.

```
nostromo remove cmd [key.path] [flags]
```

## Options

```
  -h, --help   help for cmd
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo remove](nostromo_remove.md)	 - Remove command or substitution

