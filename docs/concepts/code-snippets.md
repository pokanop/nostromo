---
title: Code snippets
---

# Code snippets

In place of a shell command, a `nostromo` command can run a one-line **code snippet** in one of these languages:

| Language | Runs                  |
| -------- | --------------------- |
| `ruby`   | `ruby -e '<snippet>'` |
| `python` | `python -c '<snippet>'` |
| `perl`   | `perl -e '<snippet>'` |
| `js`     | `node -e '<snippet>'` |
| `sh`     | the snippet as a plain shell command (the default) |

The interpreter must be on your `PATH` under that name (`python`, `node`, ...).

## Adding a snippet

Use `--code` / `-c` together with `--language` / `-l`:

```sh
nostromo add cmd hello --code 'console.log("hello js")' --language js
```

```sh
$ hello
hello js
```

```sh
nostromo add cmd py.version -l python -c 'import sys; print(sys.version)'
```

Interactive mode (`nostromo add`) also asks for a language and snippet.

## How snippets are stored and run

The snippet lives in the command's `code` block in the manifest:

```yaml
commands:
  hello:
    keypath: hello
    name: ''
    alias: hello
    code:
      language: js
      snippet: console.log("hello js")
```

When the command has a valid snippet (both `language` and `snippet` set), it takes the place of the command's `name` when the command line is built, and `nostromo eval` wraps the result in the interpreter call for the language of the command you ran.

!!! warning "Quoting"
    The snippet is wrapped in single quotes on the command line, so single quotes inside the snippet will break it. Prefer double quotes inside snippets, e.g. `print("hi")` rather than `print('hi')`.

## Multi-line snippets

The CLI only takes single-line snippets. For anything longer, edit the manifest directly (`~/.nostromo/ships/manifest.yaml` — the [`edit` example](../examples.md#edit) adds an `edit nostromo` shortcut for exactly this) and use a YAML block scalar:

```yaml
    code:
      language: python
      snippet: |
        import os
        for name in sorted(os.listdir(".")):
            print(name)
```

Multi-line YAML must be escaped correctly to work — keep the quoting caveat above in mind, since the whole snippet still ends up inside `python -c '...'`.
