# Miko-Shell CLI Tool Plan

I need you to create a GO project in this directory that will be a CLI tool

- The tool will be called `miko-shell`
- This tool serves to abstract dependencies used in a local development project and use containers
- It will also allow creating a container image based on this configuration, to be used in pipelines
- It will allow connecting to this container to execute scripts in the project context

It will have several subcommands:

- init: will create a configuration file (called dev-config.yaml) with the following format in the current directory:

```yaml
name: my-project
container-provider: docker
image: alpine:latest
pre-install:
  - apk add yamllint
shell:
  init-hook:
    - printf "Hello"
  scripts:
    - name: lint
      commands:
        - yamllint -c yamllint.config .
```

Configuration file format explanation:

- name: project name used for container image naming (auto-generated from directory name, normalized)
- container-provider: the container provider to use (docker or podman). Default: docker
- image: the image to use in the project
- pre-install: list of commands to install packages
- shell: shell configuration options

  - init-hook: list of commands to execute before opening the shell
  - scripts: List of scripts available for execution in the container context. Scripts have the following structure:
    - name: command name
    - commands: list of commands to execute in that script

- build: this command will generate a container image from the configuration

  - The base image will be the one specified in the `image` field
  - It will install packages specified in `pre-install`
  - The image name will be: the value of the `name` field in the configuration
  - The image tag will be generated from the `hash` of the `dev-config.yaml` file

- run: this command will allow executing commands inside the container

  - The container image used will be called by the `name` field and the tag will be the hash of the `dev-config` file
    - If this container doesn't exist, it will build first and then use it
  - The container will start, execute the command and terminate
  - This container will have as current directory, the project directory (where the dev-config file is)
  - All container outputs, stdin, stdout and stderr, will be directed to the shell calling this command
  - The first thing it will do before executing any command is execute the list of commands specified in shell/init-hook
  - If the command name matches the name of a script specified in shell/scripts, it will execute this script, otherwise it will execute the command in the container

- shell: This command will allow accessing the container shell
  - The container image used will be called by the `name` field and the tag will be the hash of the `dev-config` file
    - If this container doesn't exist, it will build first and then use it
  - This container will have as current directory, the project directory (where the dev-config file is)
  - The first thing it will do before executing any command is execute the list of commands specified in shell/init-hook
  - Then the user will access the container shell

## Additional Requirements:

- The tool requires Docker or Podman installed and accessible on the system
- The project directory will be mounted as a volume at `/workspace` inside the container
- The working directory inside the container will be `/workspace`
- If the `dev-config.yaml` file doesn't exist, the `build`, `run` and `shell` commands will show an error indicating that `miko-shell init` must be executed first
- The tool will include a `help` command that will show usage information

## Additional Commands:

- help: will show the tool's help with all available commands
- version: will show the current version of the tool

## Implementation Details:

- The `dev-config.yaml` file hash will be calculated using SHA256
- Containers will be executed with the `--rm` flag for automatic cleanup
- The `cobra` library will be used for CLI command handling
- The `viper` library will be used for YAML configuration
- Proper error handling with descriptive messages
- Optional logging for debugging
- Support for both Docker and Podman as container providers
- Container provider detection and validation
