---
title: find
---

# nostromo find

Find commands and substitutions by name or key path

## Synopsis

Find matching commands and substitutions in nostromo.

If the argument is the exact key path of a command in any docked manifest,
only that command and its substitutions are printed:

	nostromo find dev.install.goenv

Otherwise every command whose name, alias or key path contains the argument
is printed along with any matching substitutions:

	nostromo find install

Use --verbose to print results as tables including the manifest each match
belongs to. Use --exact to require an exact key path match and --all to list
every match even when the argument is itself a key path.

```
nostromo find [name|key.path] [flags]
```

## Options

```
  -a, --all     List all matches even when the argument is an exact key path
  -e, --exact   Only match a command at exactly the given key path
  -h, --help    help for find
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

