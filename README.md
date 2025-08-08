# Miko Shell 🐚

_Consistent environments ## 3 – Open a dev shell
miko-shell open – Open a dev shell
miko-shell openthout the clutter_

[![CI](https://github.com/jepemo/miko-shell/actions/workflows/ci.yml/badge.svg)](https://github.com/jepemo/miko-shell/actions/workflows/ci.yml)
[![Release](https://github.com/jepemo/miko-shell/actions/workflows/release.yml/badge.svg)](https://github.com/jepemo/miko-shell/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jepemo/miko-shell)](https://goreportcard.com/report/github.com/jepemo/miko-shell)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Miko Shell packages your project inside a lightweight container so every developer—local, remote or CI—runs the same toolchain. Simple, repeatable, and easy to clean up.

---

## 🚀 Why use it

- **Single YAML setup** – replaces long “install X, Y, Z” guides.
- **Docker _or_ Podman** – choose the engine you already use.
- **Fast rebuilds** – layer caching avoids redundant work.
- **Cross‑platform** – Linux, macOS, Windows (WSL 2).
- **Named scripts** – run `miko-shell run test` instead of memorising commands.
- **No lock‑in** – underneath it’s just containers.

---

## 🛠 Installation (under a minute)

| Option               | Command                                                                          | When to use                                            |
| -------------------- | -------------------------------------------------------------------------------- | ------------------------------------------------------ |
| **Bootstrap script** | `./bootstrap.sh`                                                                 | Fresh checkout; downloads Go 1.23.4, builds and tests. |
| **Go toolchain**     | `make build`<br/>or `go build -o miko-shell .`                                   | If you’re working on Miko Shell itself.                |
| **Pre‑built binary** | Download from the [releases page](https://github.com/jepemo/miko-shell/releases) | CI pipelines or quick evaluation.                      |

---

## Quick‑start

```bash
# 1 – Initialise in your project folder
miko-shell init            # or   miko-shell init --dockerfile

# 2 – Edit the generated miko-shell.yaml (example below)

# 3 – Open a dev shell
miko-shell shell

# 4 – Run predefined tasks
miko-shell run test
miko-shell run greet Alice 42
```

### Minimal example `miko-shell.yaml`

```yaml
name: my-project
container:
  provider: docker
  image: alpine:latest
  setup:
    - apk add curl git
shell:
  startup:
    - echo "Welcome to the development shell"
  scripts:
    - name: test
      commands:
        - go test ./...
    - name: greet
      commands:
        - echo "Hello $1, you are $2 years old"
```

---

## 🧩 Configuration overview

| Key                  | Purpose                                                                                         |
| -------------------- | ----------------------------------------------------------------------------------------------- |
| `name`               | Project label; also used as an image tag.                                                       |
| `container.provider` | `docker` (default) or `podman`.                                                                 |
| `container.image`    | Base image if you **don’t** supply a Dockerfile.                                                |
| `container.build.*`  | Path, context and build args when you **do** use a Dockerfile.                                  |
| `container.setup`    | Commands executed at build time (apt‑get, apk, npm…).                                           |
| `shell.startup`      | Commands executed on every `shell` or `run`.                                                    |
| `shell.scripts[]`    | Named tasks; call with `miko-shell run <name> [args…]`. Positional `$1`, `$2`, … are available. |

---

## 🍱 Templates to start from

Ready‑to‑use configurations live in `examples/`:

Python • Node/Pnpm • Go • Rust • Elixir/Phoenix • PHP • Ruby/Rails • Java
Copy, rename, tweak:

```bash
cp examples/miko-shell-go.example.yaml miko-shell.yaml
```

---

## ⚙️  How it works internally

1. **Reads** your `miko-shell.yaml`.
2. **Builds or reuses** an image tagged `name:<config‑hash>`.
3. **Mounts** the repository at `/workspace`.
4. **Executes** the script or opens an interactive shell.
5. **Leaves** your host system untouched.

---

## 🔎 Tips

- **Ad‑hoc commands**: `miko-shell run -- cargo tree` – everything after `--` is passed verbatim.
- **Performance**: keep heavy dependency installation in `setup`; `startup` runs on every command.
- **CI integration**: configure your pipeline to run `miko-shell run test` instead of duplicating logic.
- **Switch engines**: change `provider` to `podman` if that’s what your team prefers.

---

## 🤝 Contributing

1. Fork the repository.
2. `git switch -c feature/your-idea`
3. Code, commit, test.
4. Open a pull request.

Clear commit messages help reviewers and future developers.

---

## ⚖️ License

MIT License.

---

Questions or issues? Open a discussion or report a bug at [https://github.com/jepemo/miko-shell](https://github.com/jepemo/miko-shell). We appreciate your feedback.
