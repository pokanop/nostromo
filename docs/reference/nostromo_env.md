---
title: env
---

# nostromo env

Print the effective environment for a command

## Synopsis

Print the effective environment for a command in nostromo.

Merges the env vars and .env files of the command at "key.path" with those
inherited from its parents, in the order they are exported when the command
runs. Vars print as KEY=VALUE lines, use -v to see where each was defined or
-s to print export statements for a shell:
  nostromo env foo.bar
  nostromo env foo.bar -v
  nostromo env foo.bar -s fish

```
nostromo env [key.path] [flags]
```

## Options

```
  -h, --help           help for env
  -s, --shell string   Print export statements for shell (bash, zsh, fish, powershell)
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

