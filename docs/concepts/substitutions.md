---
title: Substitutions
---

# Substitutions

A **substitution** is a short alias for a longer argument. Substitutions are attached to a command scope in the tree, and `nostromo` swaps them into the arguments you type at execution time — for that command *and every command beneath it*.

## Adding a substitution

```sh
nostromo add sub <key.path> <original> <alias>
```

For example, if you frequently pass a long path to `foo bar`:

```sh
nostromo add sub foo.bar //some/long/string sls
```

Subsequent calls to `foo bar` (or `foo bar baz`, since `baz` is inside `foo.bar`'s scope) replace `sls` before running:

```sh
foo bar baz sls
```

results in:

```sh
foo bar baz //some/long/string
```

Substitutions are stored on the command node in the manifest:

```yaml
commands:
  foo:
    commands:
      bar:
        subs:
          sls:
            name: //some/long/string
            alias: sls
```

## Scope and precedence

- A substitution is visible to the command it's added to and all of its descendants. Add it at the highest node that makes sense so every child benefits.
- If the same alias is defined at several levels, the **closest** one to the command you ran wins (the tree is walked from the leaf back to the root).
- Substitutions only apply to arguments you type after the keypath, never to the command text itself. Every argument is checked, so several substitutions can be used in one call.
- An argument that doesn't match any substitution is passed through untouched.

## Substitutions and `$1`

Arguments are substituted *before* they're inserted into `$1`, `$2`, … placeholders or appended to the command, so the two features compose. See [Arguments](commands.md#arguments).

## Removing and finding

```sh
nostromo remove sub foo.bar sls   # remove by key path and alias
nostromo find sls                 # search substitutions (and commands) by name
```

Run `nostromo add` without arguments and choose **substitution** for an interactive walkthrough.

## Example

Shorten Kubernetes namespaces for a whole `k` command tree:

```sh
nostromo add cmd k "kubectl"
nostromo add cmd k.pods "get pods -n"
nostromo add cmd k.logs "logs -n"
nostromo add sub k my-very-long-production-namespace prod
nostromo add sub k my-very-long-staging-namespace stg

k pods prod              # runs: kubectl get pods -n my-very-long-production-namespace
k logs stg api-7d9f      # runs: kubectl logs -n my-very-long-staging-namespace api-7d9f
```
