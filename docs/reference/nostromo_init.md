---
title: init
---

# nostromo init

Initialize nostromo configuration

## Synopsis

Create a nostromo config file with defaults.

By default the config file is located at ~/.nostromo/ships/manifest.yaml.

Customize this with the $NOSTROMO_HOME environment variable

A welcome banner is shown the first time a config is created when running
in a terminal. Pass --no-banner to suppress it (e.g. in scripts).

```
nostromo init [flags]
```

## Options

```
  -h, --help        help for init
      --no-banner   do not print the welcome banner
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

