---
title: update
---

# nostromo update

Update a command in nostromo manifest

## Synopsis

Update a command in nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

This will update appropriate command scopes for all levels in the provided
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

A command's platforms can be changed with -p or --platforms using Go OS names
(e.g., linux, darwin, windows) or OS/arch pairs (e.g., linux/arm64). Existing
platforms are kept when the flag is omitted, pass an empty list to allow all:

    nostromo update foo.bar --platforms linux,darwin
    nostromo update foo.bar --platforms ""

```
nostromo update [key.path] [command] [options] [flags]
```

## Options

```
  -a, --alias-only           Add shell alias only, not a nostromo command
  -c, --code string          Code snippet to run for this command
  -d, --description string   Description of the command to update
  -h, --help                 help for update
  -l, --language string      Language of code snippet (e.g., ruby, python, perl, js)
  -m, --mode string          Set the mode for the command (concatenate, independent, exclusive)
  -p, --platforms strings    Limit the command to platforms (e.g., linux,darwin,windows/arm64), empty for all
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

