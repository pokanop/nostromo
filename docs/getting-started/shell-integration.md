---
title: Shell integration
---

# Shell integration

`nostromo` supports `bash`, `zsh`, `fish` and PowerShell. `nostromo init` adds a block to the startup files that **already exist** for these shells:

```sh
# nostromo [section begin]
source <(nostromo completion bash)
# nostromo [section end]
```

The block sources the output of `nostromo completion <shell>`, which contains:

- a `nostromo` wrapper function that re-sources completions after every successful `nostromo` command, so new commands are available immediately;
- a shell function (or plain alias for [alias-only](../concepts/alias-only.md) commands) for every top level command in your manifests;
- tab completion for `nostromo` itself and for every command you add, including their descriptions.

| Shell      | Startup file                                                                                                         | Block                                                                   |
| ---------- | -------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| bash       | `~/.bashrc`                                                                                                          | `source <(nostromo completion bash)`                                    |
| zsh        | `~/.zshrc` (or `$ZDOTDIR/.zshrc`)                                                                                    | `autoload -U compinit; compinit`<br>`source <(nostromo completion zsh)` |
| fish       | `~/.config/fish/config.fish` (or `$XDG_CONFIG_HOME/fish/config.fish`)                                                | `nostromo completion fish | source`                                    |
| PowerShell | `~/.config/powershell/Microsoft.PowerShell_profile.ps1` (Windows: `~/Documents/PowerShell/Microsoft.PowerShell_profile.ps1`) | `nostromo completion powershell | Out-String | Invoke-Expression`     |

!!! note "The startup file must exist"
    `nostromo` only edits files that already exist so it never creates a startup file for a shell you don't use. If you use zsh but only have a `~/.bashrc`, create `~/.zshrc` first (`touch ~/.zshrc`) and run `nostromo init` again. Before writing, `nostromo` saves a timestamped backup of the startup file to your system temp directory.

Legacy `~/.profile` and `~/.bash_profile` files are also scanned: an existing `nostromo` block in them is kept up to date but no new block is added.

## Manual setup

If you'd rather not have `nostromo` touch your startup file, or want to load completions from a static file, `nostromo completion` prints the script for any supported shell:

=== "bash"

    ```sh
    source <(nostromo completion bash)

    # To load completions for each session, execute once:
    # Linux:
    nostromo completion bash > /etc/bash_completion.d/nostromo
    # macOS:
    nostromo completion bash > /usr/local/etc/bash_completion.d/nostromo
    ```

=== "zsh"

    ```sh
    # If shell completion is not already enabled in your environment,
    # you will need to enable it. You can execute the following once:
    echo "autoload -U compinit; compinit" >> ~/.zshrc

    # To load completions for each session, execute once:
    nostromo completion zsh > "${fpath[1]}/_nostromo"

    # You will need to start a new shell for this setup to take effect.
    ```

=== "fish"

    ```sh
    nostromo completion fish | source

    # To load completions for each session, execute once:
    nostromo completion fish > ~/.config/fish/completions/nostromo.fish
    ```

=== "PowerShell"

    ```powershell
    nostromo completion powershell | Out-String | Invoke-Expression

    # To load completions for every new session, run:
    nostromo completion powershell > nostromo.ps1
    # and source this file from your PowerShell profile.
    ```

!!! warning
    Writing the completion script to a static file means it won't include commands you add later. Regenerate the file (or use the `source <(nostromo completion ...)` form) after changing your manifests.

`nostromo init` also writes the scripts for all four shells to `~/.nostromo/completions/nostromo.{bash,zsh,fish,powershell}`.

## What gets generated

For a manifest with a `foo` command and an alias-only `c` command, the bash script contains function definitions along these lines:

```sh
__nostromo_cmd() { command nostromo "$@"; }
nostromo() { __nostromo_cmd "$@" && eval "$(__nostromo_cmd completion bash)"; }

alias c='clear'
foo() { eval $(__nostromo_cmd eval foo "$@"); }
```

The same definitions are emitted in fish and PowerShell syntax for those shells, so your commands behave identically everywhere.

## Man pages

`nostromo init` generates man pages under `~/.nostromo/man` and attempts to symlink them into `/usr/local/share/man/man1` (skipped on Windows), so `man nostromo-add-cmd` works too. If that folder isn't writable the links are silently skipped; the pages are still available under `~/.nostromo/man`.

## Verify

```sh
nostromo show
```

prints your manifests, config and commands, and ends with a `[profile]` section showing the block found in your startup file. If it's missing, check that the file for your shell exists and re-run `nostromo init`.
