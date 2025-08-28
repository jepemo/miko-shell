# Miko Shell

[![Release](https://img.shields.io/github/v/release/jepemo/miko-shell)](https://github.com/jepemo/miko-shell/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/jepemo/miko-shell/ci.yml)](https://github.com/jepemo/miko-shell/actions)
[![Go Report](https://goreportcard.com/badge/github.com/jepemo/miko-shell)](https://goreportcard.com/report/github.com/jepemo/miko-shell)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jepemo/miko-shell)](https://github.com/jepemo/miko-shell/blob/main/go.mod)
[![License](https://img.shields.io/github/license/jepemo/miko-shell)](LICENSE)

Declarative, reproducible dev environments backed by Docker or Podman. One YAML file, same toolchain for every developer and CI job.

---

## Quick Start

```bash
# 1) Create a config
miko-shell init               # or: miko-shell init --dockerfile

# 2) (Optional) Build the image
miko-shell build

# 3) Discover available scripts
miko-shell run

# 4) Run a script
miko-shell run test
```

Minimal `miko-shell.yaml`:

```yaml
name: my-project
container:
  provider: docker
  image: alpine:latest
  setup:
    - apk add --no-cache curl git
shell:
  startup:
    - echo "Welcome to the development shell"
  scripts:
    - name: test
      commands:
        - go test ./...
```

---

## Install

- Quick install:

```bash
curl -sSL https://raw.githubusercontent.com/jepemo/miko-shell/main/install.sh | bash
```

- Uninstall:

```bash
curl -sSL https://raw.githubusercontent.com/jepemo/miko-shell/main/install.sh | bash -s -- --uninstall
```

- Bootstrap from local checkout: `./bootstrap.sh`
- From source: `make build` or `go build -o miko-shell .`
- Prebuilt binaries: see [Releases](https://github.com/jepemo/miko-shell/releases)

---

## Commands (condensed)

- `init` — scaffold a config (`--dockerfile` for Dockerfile-based builds)
- `build` — build the container image from the config
- `run` — list scripts (no args) or run `run <name> [args...]`
- `version` — print version

For details and advanced usage, see [DOCS.md](DOCS.md).

---

## Examples

Start from ready-made configs in `examples/`:

```bash
# Go
miko-shell build -c examples/dev-config-go.example.yaml
miko-shell run   -c examples/dev-config-go.example.yaml test

# Next.js
miko-shell build -c examples/dev-config-nextjs.example.yaml
miko-shell run   -c examples/dev-config-nextjs.example.yaml dev
```

More examples and tips: `examples/README.md`, `examples/USAGE.md`.

---

## License

MIT — see [LICENSE](LICENSE).
