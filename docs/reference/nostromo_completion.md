---
title: completion
---

# nostromo completion

Generate completion script

## Synopsis

To load completions:

Bash:

  $ source <(nostromo completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ nostromo completion bash > /etc/bash_completion.d/nostromo
  # macOS:
  $ nostromo completion bash > /usr/local/etc/bash_completion.d/nostromo

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ nostromo completion zsh > "${fpath[1]}/_nostromo"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ nostromo completion fish | source

  # To load completions for each session, execute once:
  $ nostromo completion fish > ~/.config/fish/completions/nostromo.fish

PowerShell:

  PS> nostromo completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> nostromo completion powershell > nostromo.ps1
  # and source this file from your PowerShell profile.


```
nostromo completion [bash|zsh|fish|powershell]
```

## Options

```
  -h, --help   help for completion
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

