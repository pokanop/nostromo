---
title: Commands and keypaths
---

# Commands and keypaths

Aliases — or **commands** in `nostromo` parlance — are the core of `nostromo`. Instead of constantly updating shell profiles by hand, `nostromo` keeps your shell up to date with the latest additions automatically.

!!! info "How commands run"
    `nostromo` is not a shell builtin, so a couple of things are worth knowing:

    - Commands are resolved by the `nostromo` binary and executed with `eval` inside a shell function, so they behave like aliases (`cd` persists, environment variables stick).
    - Commands and changes are available immediately because `nostromo` reloads completions automatically after every command.

    If you just want a boring standard shell alias you can have that too — see [Alias-only commands](alias-only.md).

## Adding a command

```sh
nostromo add cmd foo "echo bar"
```

And just like that you can run `foo` like any other alias. Descriptions show up in tab completion, so they're worth adding:

```sh
nostromo add cmd foo "echo bar" -d "My magical foo command that prints bar"
```

`add cmd` accepts these flags (see [`nostromo add cmd`](../reference/nostromo_add_cmd.md)):

| Flag                  | Purpose                                                                  |
| --------------------- | ------------------------------------------------------------------------ |
| `-d, --description`   | Description shown in `nostromo show` and in tab completion               |
| `-m, --mode`          | [Execution mode](modes.md): `concatenate`, `independent` or `exclusive` |
| `-a, --alias-only`    | Create a plain shell [alias](alias-only.md) instead of a `nostromo` command |
| `-c, --code`          | A [code snippet](code-snippets.md) to run instead of a shell command     |
| `-l, --language`      | Language of the snippet: `ruby`, `python`, `perl` or `js`                |

Run `nostromo add` with no arguments for an interactive walkthrough.

## Keypaths

`nostromo` uses **keypaths** to build and address the command tree. A keypath is a `.` delimited string that represents the path to a command. For example:

```sh
nostromo add cmd foo.bar.baz 'echo hello'
```

builds the tree `foo` :material-arrow-right: `bar` :material-arrow-right: `baz`, so that any of these are now valid:

```sh
foo
foo bar
foo bar baz
```

Only the last one runs `echo hello` — `foo` and `foo bar` were created as empty intermediate nodes. You can give them commands too by adding at those keypaths:

```sh
nostromo add cmd foo 'echo oof'
nostromo add cmd foo.bar 'rab'
```

Now `foo bar baz` runs `echo oof rab hello`. Walking the tree, each node's command is added to the final command line — the **default** [mode](modes.md) is to concatenate. Targeted use of `;` or `&&` inside your commands lets you run several things instead of concatenating, or you can switch a command's mode to `independent` to have `nostromo` do it for you.

!!! note
    Keypath segments become the words you type in your shell, so avoid `.` and whitespace in aliases. `-` and `_` are fine (`ls-la`, `bash_profile`).

## Arguments

Anything you type after the keypath is passed along to the command. By default the extra arguments are appended to the end of the resolved command line:

```sh
nostromo add cmd gco "git checkout"
gco -b feature   # runs: git checkout -b feature
```

If the command contains positional placeholders `$1`, `$2`, … they are replaced with the corresponding arguments instead, and only the leftovers are appended:

```sh
nostromo add cmd check 'which $1 > /dev/null 2>&1 && echo $1 exists || (echo $1 not found && exit 1)'
check jq         # runs the command with every $1 replaced by jq
```

Arguments also go through [substitutions](substitutions.md) before being inserted.

## Inspecting and searching

```sh
nostromo show            # manifests, config, commands and your profile block
nostromo show --tree     # the command tree
nostromo find bar        # commands and substitutions whose name matches "bar"
nostromo eval foo bar    # print the command that would run, without running it
```

`eval` is what the generated shell functions call under the hood, so it's the quickest way to see exactly how a keypath resolves.

## Updating and removing

Update any property of an existing command with the same flags as `add cmd`:

```sh
nostromo update foo.bar "echo rab" -d "prints rab" -m independent
```

Remove a command and everything beneath it:

```sh
nostromo remove cmd foo.bar
```

For moving, copying, renaming and splitting subtrees see [Tree management](tree-management.md).

## Disabling a command

A command node has a `disabled` flag in the manifest YAML. There's no CLI switch for it, but setting `disabled: true` on a node by hand disables that command *and every command beneath it*; running one reports `command is disabled at <keypath>`.
