---
title: Installation
---

# Installation

## Homebrew (macOS and Linux)

```sh
brew install pokanop/pokanop/nostromo
```

The formula lives in the [pokanop/homebrew-pokanop](https://github.com/pokanop/homebrew-pokanop) tap and is updated automatically on every release.

## Go

With Go 1.17 or newer:

```sh
go install github.com/pokanop/nostromo@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`.

## Pre-built binaries

Every [GitHub release](https://github.com/pokanop/nostromo/releases) ships archives for Linux, macOS and Windows named `nostromo_<Os>_<Arch>.tar.gz` (or `.zip` on Windows), e.g. `nostromo_Darwin_arm64.tar.gz` or `nostromo_Linux_x86_64.tar.gz`, along with a `checksums.txt`. Download the archive for your platform, extract it and put the `nostromo` binary somewhere on your `PATH`.

## From source

```sh
git clone https://github.com/pokanop/nostromo.git
cd nostromo
go build -o ~/bin/nostromo .
```

## Verify

```sh
nostromo version
```

prints the tag, commit hash and build date of your install.

## Upgrading

Upgrade with the same tool you installed with (`brew upgrade nostromo`, `go install ...@latest`, or download a new archive). After upgrading, run:

```sh
nostromo init
```

`nostromo init` is safe to run repeatedly. It migrates any manifest changes, regenerates completion scripts and man pages, and makes sure your shell startup file still has the `nostromo` block.

Next: the [Quick start](quickstart.md).
