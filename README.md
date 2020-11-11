# Om

_a mantra, or vibration, that is traditionally chanted_

![enhancing your calm](http://i.giphy.com/3o7qDQ5iw1oXyDeJAk.gif)

`om` is a tool that helps you configure and deploy tiles to Ops-Manager.
Currently being maintained by the PCF Platform Automation team,
with engineering support and review from PCF Release Engineering.
The (private) backlog for Platform Automation is [here](https://www.pivotaltracker.com/n/projects/1472134).

## Documentation

See [here](docs/README.md) for useful examples and documentation.

When working with `om`,
it can sometimes be useful to reference the Ops Manager API docs.
You can find them at
`https://pcf.your-ops-manager.example.com/docs`.

**NOTE**: Additional documentation for om commands
used in Pivotal Platform Automation
can be found in [Pivotal Documentation](https://docs.pivotal.io/platform-automation).

## Versioning

`om` went 1.0.0 on May 7, 2019

As of that release, `om` is [semantically versioned](https://semver.org/).
When consumig `om` in your CI system,
it is now safe to pin to a particular minor version line (major.minor.patch)
without fear of breaking changes.

### API Declaration for Semver

The `om` API consists of:

1. All `om` commands not marked **EXPERIMENTAL**, and the flags for those commands
1. All global `om` flags
1. The format for any inputs to non-experimental `om` commands.
1. The format for any outputs from non-experimental `om` commands.
1. The file filename of the Github relases.

"**EXPERIMENTAL**" commands are still in development.
We may rename them, alter their flags or behavior, or remove them entirely.
When we're confident a command has the right basic shape and behavior,
we'll remove the "**EXPERIMENTAL**" designation.

Changes internal to `om` will _**NOT**_ be included as a part of the om API.
The versioning here is for the CLI tool,
not any libraries or packages included therein.
The `om` team may change any internal structs, interfaces, etc,
without reflecting such changes in the version,
so long as the outputs and behavior of the commands remain the same.

Though `om` is used heavily in [Platform Automation for PCF](network.pivotal.io/platform-automation),
which is also semantically versioned.
`om` is versioned independently from `platform-automation`.

## Installation

To download `om` go to [Releases](https://github.com/pivotal-cf/om/releases).

Alternatively, you can install `om` via `apt-get`
via [Stark and Wayne](https://www.starkandwayne.com/):

```sh
# apt-get:
sudo wget -q -O - https://raw.githubusercontent.com/starkandwayne/homebrew-cf/master/public.key | sudo  apt-key add -
sudo echo "deb http://apt.starkandwayne.com stable main" | sudo  tee /etc/apt/sources.list.d/starkandwayne.list
sudo apt-get update
sudo apt-get install om -y
```


Or by the Linux and Mac `brew`

```
brew tap pivotal-cf/om https://github.com/pivotal-cf/om
brew install om
```

You can also build from source.

### Building from Source

You'll need at least Go 1.12, as
`om` uses Go Modules to manage dependencies.

To build from source, after you've cloned the repo,
run these commands from the top level of the repo:

```bash
GO112MODULE=on go mod download
GO112MODULE=on go build
```

Go 1.11 uses some heuristics to determine if Go Modules should be used.
The process above overrides those heuristics
to ensure that Go Modules are _always_ used.
If you have cloned this repo outside of your GOPATH,
`GO111MODULE=on` can be excluded from the above steps.
