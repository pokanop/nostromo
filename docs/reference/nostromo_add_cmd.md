---
title: add cmd
---

# nostromo add cmd

Add a command to nostromo manifest

## Synopsis

Add a command to nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

This will create appropriate command scopes for all levels in the provided
key path. A command scope can contain a tree of sub commands and 
substitutions.

A command's mode indicates how it will be executed. By default, nostromo
concatenates parent and child commands along the tree. There are 3 modes
available to commands:

    concatenate  Concatenate this command with subcommands exactly as defined
    independent  Execute this command with subcommands using ';' to separate
    exclusive    Execute this and only this command ignoring parent commands

You can set using -m or --mode when adding a command or globally using:

    nostromo set mode <mode>

```
nostromo add cmd [key.path] [command] [options] [flags]
```

## Options

```
  -a, --alias-only           Add shell alias only, not a nostromo command
  -c, --code string          Code snippet to run for this command
  -d, --description string   Description of the command to add
  -h, --help                 help for cmd
  -l, --language string      Language of code snippet (e.g., ruby, python, perl, js)
  -m, --mode string          Set the mode for the command (concatenate, independent, exclusive)
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo add](nostromo_add.md)	 - Add command or substitution

