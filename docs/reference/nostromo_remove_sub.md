---
title: remove sub
---

# nostromo remove sub

Remove a substitution from nostromo manifest

## Synopsis

Remove a substitution from nostromo manifest for a given key path and arg.
A substitution allows any arguments as part of a command to be substituted
by that value, e.g., "original-cmd //some/long/arg1 //some/long/arg2" can
be substituted out with "cmd-alias sub1 sub2" by adding subs for the args.

This will remove the substitution for scopes beneath levels in
the provided key path. A command scope can a tree of sub commands
and substitutions.

```
nostromo remove sub [key.path] [alias] [flags]
```

## Options

```
  -h, --help   help for sub
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo remove](nostromo_remove.md)	 - Remove command or substitution

