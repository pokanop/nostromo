---
title: Examples
---

# Examples

The [`examples/`](https://github.com/pokanop/nostromo/tree/main/examples) folder in the repository contains ready-to-use manifests. Dock any of them straight from GitHub:

```sh
nostromo dock https://github.com/pokanop/nostromo/raw/main/examples/docker.yaml
```

or all at once:

```sh
nostromo dock github.com/pokanop/nostromo//examples
```

!!! note
    `examples/manifest.yaml` is named `manifest`, the same as your core manifest. When docked it's renamed to `manifest-<timestamp>` so it doesn't clash; the other three are added under their own names. Undock anything you don't want with `nostromo undock <name>`.

## docker

[`examples/docker.yaml`](https://github.com/pokanop/nostromo/blob/main/examples/docker.yaml) — a `dock` command tree wrapping `docker` for cleaning up containers, images and volumes. A textbook use of [concatenate mode](concepts/modes.md#concatenate): the root node supplies `docker` and each child supplies the rest.

| Command       | Runs                                                   |
| ------------- | ------------------------------------------------------ |
| `dock`        | `docker`                                               |
| `dock clean`  | `docker ps -a -q \|xargs docker rm`                    |
| `dock rmi`    | `docker rmi $(docker images -q)`                       |
| `dock stop`   | `docker ps -q \|xargs docker stop`                     |
| `dock vol`    | `docker volume rm $(docker volume ls -qf dangling=true)` |

Recreate it yourself:

```sh
nostromo add cmd dock docker -d "Alias for docker"
nostromo add cmd dock.clean "ps -a -q |xargs docker rm" -d "Remove Docker containers using docker rm"
nostromo add cmd dock.rmi 'rmi $(docker images -q)' -d "Remove Docker images using docker rmi"
nostromo add cmd dock.stop "ps -q |xargs docker stop" -d "Stop running Docker containers"
nostromo add cmd dock.vol 'volume rm $(docker volume ls -qf dangling=true)' -d "Remove dangling Docker volumes"
```

## edit

[`examples/edit.yaml`](https://github.com/pokanop/nostromo/blob/main/examples/edit.yaml) — `edit <thing>` opens common dotfiles in `vim`. The root `edit` node has an **empty** command, so it contributes nothing to the command line and acts purely as a namespace.

| Command             | Runs                                  |
| ------------------- | ------------------------------------- |
| `edit bash`         | `vim ~/.bashrc`                       |
| `edit bash_profile` | `vim ~/.bash_profile`                 |
| `edit nostromo`     | `vim ~/.nostromo/ships/manifest.yaml` |
| `edit p10k`         | `vim ~/.p10k.zsh`                     |
| `edit profile`      | `vim ~/.profile`                      |
| `edit ssh`          | `vim ~/.ssh/config`                   |
| `edit ssh hosts`    | `vim ~/.ssh/hosts`                    |
| `edit terminalizer` | `vim ~/.terminalizer/config.yml`      |
| `edit zsh`          | `vim ~/.zshrc`                        |

`edit ssh hosts` is worth a closer look: it's a child of `edit ssh` but uses [exclusive mode](concepts/modes.md#exclusive) (`mode: 2`), so running it executes only `vim ~/.ssh/hosts` rather than `vim ~/.ssh/config vim ~/.ssh/hosts`.

## tools

[`examples/tools.yaml`](https://github.com/pokanop/nostromo/blob/main/examples/tools.yaml) — helpers for checking and installing tooling, mostly Homebrew based.

| Command              | Runs                                                                   |
| -------------------- | ---------------------------------------------------------------------- |
| `check <tool>`       | `which $1 > /dev/null 2>&1 && echo $1 exists \|\| (echo $1 not found && exit 1)` |
| `exists <path>`      | `[ -f $1 ] \|\| [ -d $1 ] && echo $1 exists \|\| (echo $1 not found && exit 1)` |
| `install brew`       | the official Homebrew install script                                   |
| `install jq`         | `brew install jq`                                                      |
| `install nvm`        | the nvm install script                                                 |
| `install ohmyzsh`    | the oh-my-zsh install script                                           |
| `install p10k`       | clone powerlevel10k into your oh-my-zsh custom themes                  |
| `install rust`       | `rustup` bootstrap                                                     |
| `run quietly <cmd>`  | run `$2` with output silenced and report success/failure               |

`check` and `exists` show how [positional arguments](concepts/commands.md#arguments) work: `nostromo` replaces `$1` with the first argument you typed after the command before handing the line to the shell.

```sh
$ check jq
jq exists
$ check nope
nope not found
```

## manifest

[`examples/manifest.yaml`](https://github.com/pokanop/nostromo/blob/main/examples/manifest.yaml) — a complete core manifest showing most features in one place, including the `dock` and `edit` trees above plus:

| Command          | Runs                                                        | Feature                                   |
| ---------------- | ----------------------------------------------------------- | ----------------------------------------- |
| `c`              | `clear`                                                     | [alias-only](concepts/alias-only.md) shell alias |
| `gcdf`           | `git clean -df`                                             | alias-only                                |
| `cat`            | `bat`                                                       | simple replacement                        |
| `code nostromo`  | `command code ~/.nostromo/manifest.yaml`                    | [substitutions](concepts/substitutions.md) `nostromo` and `zsh` |
| `copy ssh`       | `cat ~/.ssh/id_rsa.pub \| pbcopy`                           | namespace with empty root                 |
| `ios snap <file>`| `xcrun simctl io booted screenshot $1`                      | positional args                           |
| `ip local`       | `ifconfig en0 \| grep --word-regexp inet \| awk "{print $2}"` | pipes and quoting                       |
| `nuke docker`    | `dock stop && dock clean && dock rmi`                       | calling other `nostromo` commands         |
| `reload`         | `. ~/.zshrc`                                                | sourcing in the current shell             |

The `code` command uses `command code` rather than plain `code` — the `command` builtin makes sure the real VS Code binary runs rather than any function or alias named `code`, which matters when a `nostromo` command shares a name with the tool it wraps.

`nuke docker` demonstrates composition: because `dock` is itself a shell function generated by `nostromo`, other commands can call it like any other program.
