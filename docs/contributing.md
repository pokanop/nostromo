---
title: Contributing
---

# Contributing

Contributions are what makes the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a pull request

## Development setup

`nostromo` is a Go module with no build steps beyond the Go toolchain (Go 1.21 or newer):

```sh
git clone https://github.com/pokanop/nostromo.git
cd nostromo
go build -o ~/bin/nostromo .
```

Run the checks that CI runs before opening a pull request:

```sh
go vet ./...
go test -race ./...
for GOOS in linux darwin windows; do
  GOOS=$GOOS GOARCH=amd64 CGO_ENABLED=0 go build -o /dev/null .
done
```

A quick end-to-end smoke test in a throwaway home directory:

```sh
export HOME=$(mktemp -d)
touch $HOME/.bashrc
nostromo init
nostromo add cmd foo.bar "echo hi"
nostromo eval foo bar     # prints: echo hi
```

## Project layout

| Package        | Purpose                                                                                             |
| -------------- | --------------------------------------------------------------------------------------------------- |
| `cmd/`         | Cobra command definitions — flags, help text and wiring only                                          |
| `task/`        | The controller layer. Each CLI action is a small `task.*` function that returns an exit code         |
| `model/`       | `Manifest`, `Spaceport`, `Command`, `Substitution`, `Code`, `Mode` and `Config` and their behaviour   |
| `config/`      | Loading/saving YAML, paths under `~/.nostromo`, backups (`cargo`) and `dock`/`sync` downloads        |
| `shell/`       | Evaluating commands, generating shell functions/aliases, completion scripts and startup-file edits    |
| `keypath/`     | Dot-delimited keypath helpers                                                                        |
| `log/`         | Themed, levelled output                                                                              |
| `examples/`    | Sample manifests, also used by the [Examples](examples.md) page                                       |
| `docs/`        | This documentation site                                                                              |

Conventions:

- Keep `cmd/` thin: parse flags, call one `task.*` function, `os.Exit` with its result.
- Print through `log.*` (`log.Highlight`, `log.Error`, `log.Debug`, ...) rather than `fmt`, so themes and verbosity apply.
- Tests live next to the code (`foo_test.go`) and use the standard library `testing` package.
- Manifests are serialized with `gopkg.in/yaml.v2` using **lowercased field names**; existing users' manifests must keep loading, so treat the [YAML schema](schema.md) as a compatibility contract.

## Working on the documentation

The site is built with [MkDocs](https://www.mkdocs.org/) and the [Material](https://squidfunk.github.io/mkdocs-material/) theme from the `docs/` folder and `mkdocs.yml`.

```sh
python3 -m venv .venv && source .venv/bin/activate
pip install -r docs/requirements.txt
mkdocs serve
```

Then open <http://127.0.0.1:8000>. `mkdocs build --strict` fails on broken links and is what CI runs.

### Command reference

The pages under `docs/reference/` are **generated** from the Cobra command definitions — don't edit them by hand. After changing anything in `cmd/` (a flag, help text, a new command) regenerate them and commit the result:

```sh
go run . docs
```

CI checks that the committed reference matches the code (`go run . docs && git diff --exit-code docs/reference`).

### Deployment

Pushes to `main` build the site and deploy it to GitHub Pages at [nostromo.sh](https://nostromo.sh) via `.github/workflows/docs.yml`. Pull requests only build the site to catch errors.

## Releases

Releases are cut with [GoReleaser](https://goreleaser.com/) from tags (`.goreleaser.yml` and `.github/workflows/release.yml`). Homebrew users install from the `pokanop/pokanop` tap.

## License

Distributed under the MIT License.
