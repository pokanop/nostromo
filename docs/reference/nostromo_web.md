---
title: web
---

# nostromo web

Serve a local web UI for editing manifests

## Synopsis

Serve a local web UI for editing manifests.

Starts a web server bound to 127.0.0.1 and opens the UI in your browser.
The UI lists docked manifests, shows the command tree and lets you add,
edit, search and remove commands and substitutions in the core manifest.

Changes are saved through the same path as the CLI so cargo backups keep
working. Docked manifests are shown read-only since syncing overwrites them.

Press ctrl+c to stop the server.

```
nostromo web [flags]
```

## Options

```
  -h, --help       help for web
      --no-open    Do not open the browser automatically
  -p, --port int   Port to listen on, use 0 for a random free port (default 8080)
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

