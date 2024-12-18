# sd-local
[![Build Status][build-image]][build-url]
[![Latest Release][version-image]][version-url]
[![Go Report Card][goreport-image]][goreport-url]

Screwdriver local mode. See [User Guide](https://docs.screwdriver.cd/user-guide/local) for details.

## Usage

### Install

See [User Guide](https://docs.screwdriver.cd/user-guide/local) for instructions.

#### Installing locally using homebrew

Tap and install `sd-local`.

```bash
brew tap screwdriver-cd/sd-local https://github.com/screwdriver-cd/sd-local.git
brew install sd-local
```

### Execute
```bash
$ sd-local --help
Run build instantly on your local machine with
a mostly the same environment as Screwdriver.cd's

Usage:
  sd-local [command]

Available Commands:
  build       Run screwdriver build.
  config      Manage settings related to sd-local.
  help        Help about any command
  version     Display command's version.

Flags:
  -h, --help      help for sd-local
  -v, --verbose   verbose output.

Use "sd-local [command] --help" for more information about a command.
```

##### build
```bash
$ sd-local build --help
Run screwdriver build of the specified job name.

Usage:
  sd-local build [job name] [flags]

Flags:
      --artifacts-dir string   Path to the host side directory which is mounted into $SD_ARTIFACTS_DIR. (default "sd-artifacts")
  -e, --env stringToString     Set key and value relationship which is set as environment variables of Build Container. (<key>=<value>) (default [])
      --env-file string        Path to config file of environment variables. '.env' format file can be used.
  -h, --help                   help for build
  -i, --interactive            Attach the build container in interactive mode.
  -m, --memory string          Memory limit for build container, which take a positive integer, followed by a suffix of b, k, m, g.
      --meta string            Metadata to pass into the build environment, which is represented with JSON format
      --meta-file string       Path to the meta file. meta file is represented with JSON format.
      --no-image-pull          Skip container image pulls to save time.
      --privileged             Use privileged mode for container runtime.
  -S, --socket string          Path to the socket. It will used in build container.
      --src-url string         Specify the source url to build.
                               ex) git@github.com:<org>/<repo>.git[#<branch>]
                                   https://github.com/<org>/<repo>.git[#<branch>]
      --sudo                   Use sudo command for container runtime.
      --vol string             Mount local volumes into build container. (<src>:<destination>) (default [])

  -u, --user string            Change default build user. Default value is from container in use.
Global Flags:
  -v, --verbose   verbose output.
```
* There are some ways to set environment variables for a build. If the same key is set in more than one way, the priority is as follows:
  1. `--env` or `-e` Flag
  1. `--env-file` Flag
  1. environment in a screwdriver.yaml
  1. defaultEnv (e.g.: `SD_TOKEN`, `SD_API_URL`)

* You can execute step commands in interactive mode.

```bash
sd-local# sdrun --help
Run predefined steps of the specified job

Usage:
  sdrun [flags]
  sdrun [step]

Flags:
  --help        help for sdrun
  --list        show all steps
  --all         run all steps
```

* You can use docker commands in the build to run containers, build images, etc.
  * Set `screwdriver.cd/dockerEnabled: true` in the job annotations.
```yaml
jobs:
  main:
    annotations:
      screwdriver.cd/dockerEnabled: true
```

##### config
_create_
```bash
$ sd-local config create --help
Create the config of sd-local.
The new config has only launcher-version and launcher-image.

Usage:
  sd-local config create [name] [flags]

Flags:
  -h, --help   help for create

Global Flags:
  -v, --verbose   verbose output.
```

_delete_
```bash
$ sd-local config delete --help
Delete the config of sd-local.

Usage:
  sd-local config delete [name] [flags]

Flags:
  -h, --help   help for delete

Global Flags:
  -v, --verbose   verbose output.
```

_use_
```bash
$ sd-local config use --help
Use the specified config as current config.
You can confirm the current config in view sub command.

Usage:
  sd-local config use [name] [flags]

Flags:
  -h, --help   help for use

Global Flags:
  -v, --verbose   verbose output.
```

_set_
```bash
$ sd-local config set --help
Set the config of sd-local.
Can set the below settings:
* Screwdriver.cd API URL as "api-url"
* Screwdriver.cd Store URL as "store-url"
* Screwdriver.cd Token as "token"
* Screwdriver.cd launcher version as "launcher-version"
* Screwdriver.cd UUID as "uuid"
* Screwdriver.cd launcher image as "launcher-image"

Usage:
  sd-local config set [key] [value] [flags]

Flags:
  -h, --help   help for set

Global Flags:
  -v, --verbose   verbose output.
```

_view_
```bash
$ sd-local config view
  anoter:
    api-url: ""
    store-url: ""
    token: ""
    launcher:
      version: stable
      image: screwdrivercd/launcher
* default:
    api-url: https://api.screwdriver.cd
    store-url: https://store.screwdriver.cd
    token: <API Token>
    launcher:
      version: stable
      image: screwdrivercd/launcher
```

##### version
```bash
$ sd-local version
1.0.23
platform: darwin/amd64
go: go1.15.7
compiler: gc
```

##### update
```bash
$ sd-local update
Do you want to update to 1.0.5? [y/N]: y
Successfully updated to version 1.0.5
```
If you get the following error while running the update command,
```
Error occurred while detecting version: GET https://api.github.com/repos/screwdriver-cd/sd-local/releases: 403 API rate limit exceeded.
```
Please set the [GitHub personal access token.](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token)
###### bash
```bash
export GITHUB_TOKEN=<token>
```

## Testing
```bash
$ go get github.com/screwdriver-cd/sd-local
$ go test -cover github.com/screwdriver-cd/sd-local/...
```

## FAQ
1. How to forward your local `ssh-agent` to `sd-local` in Mac? 

   If your `sd-local` build needs access to ssh keys for your ssh-agent, then you can do one of the following options:
   1. If using [Colima](https://github.com/abiosoft/colima/):
  
      Make sure to start Colima with ssh-agent `colima start --ssh-agent`
      ```
      sd-local build -i docker-build-api-pr --vol "$HOME/.ssh/known_hosts:/root/.ssh/known_hosts" -S $(colima ssh eval 'echo $SSH_AUTH_SOCK')
      ```
   2. If using [Docker Desktop](https://www.docker.com/products/docker-desktop):

      Make sure to run Docker for Mac from a terminal `killall Docker && open /Applications/Docker.app`
      
      ```
      sd-local build -i docker-build-api-pr --vol "$HOME/.ssh/known_hosts:/root/.ssh/known_hosts"
      ```

## License
Code licensed under the BSD 3-Clause license. See [LICENSE](https://github.com/screwdriver-cd/sd-local/blob/master/LICENSE) file for terms.

[version-image]: https://img.shields.io/github/tag/screwdriver-cd/sd-local.svg
[version-url]: https://github.com/screwdriver-cd/sd-local/releases
[build-image]: https://cd.screwdriver.cd/pipelines/4014/badge
[build-url]: https://cd.screwdriver.cd/pipelines/4014
[goreport-image]: https://goreportcard.com/badge/github.com/Screwdriver-cd/sd-local
[goreport-url]: https://goreportcard.com/report/github.com/Screwdriver-cd/sd-local
