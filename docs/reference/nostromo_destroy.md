---
title: destroy
---

# nostromo destroy

Destroy nostromo configuration

## Synopsis

Destroy nostromo configuration and start fresh.

By default the core manifest is only destroyed and recreated.

Optionally delete the entire installation using -n flag. Note that
this does not remove shell init file entries added by nostromo. 
You'll have to delete those manually.

```
nostromo destroy [flags]
```

## Options

```
  -h, --help   help for destroy
  -n, --nuke   Nuke the entire installation
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

