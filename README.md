# yaml-docs

[![Go Report Card](https://goreportcard.com/badge/github.com/blakyaks/yaml-docs)](https://goreportcard.com/report/github.com/blakyaks/yaml-docs)

## Acknowledgements

The project was initially based on a fork of the awesome [helm-docs](https://github.com/norwoodj/helm-docs) tool, but no longer forks this
repository due to divergence in parts of the codebase.

## About

<p align="left" sytle="float: left;">
  <img alt="yaml-docs" width="25%" src="./docs/yaml-docs.png" style="float: left;"/>
</p>

**`yaml-docs`** is a tool that auto-generates documentation based on comments in YAML configuration files.

The markdown generation is entirely [gotemplate](https://golang.org/pkg/text/template) driven. The tool parses metadata
from YAML files and generates a number of sub-templates that can be referenced in a template file (by default `README.md.gotmpl`).
If no template file is provided, the tool has a default internal template that will generate a reasonably formatted README.

The most useful aspect of this tool is the auto-detection of field descriptions from comments:

```yaml
config:
  databasesToCreate:
    # -- default database for storage of database metadata
    - postgres

    # -- database for the [hashbash](https://github.com/norwoodj/hashbash-backend-go) project
    - hashbash

  usersToCreate:
    # -- admin user
    - {name: root, admin: true}

    # -- user with access to the database with the same name
    - {name: hashbash, readwriteDatabases: [hashbash]}

statefulset:
  image:
    # -- Image to use for deploying, must support an entrypoint which creates users/databases from appropriate config files
    repository: jnorwood/postgresql
    tag: "11"

  # -- Additional volumes to be mounted into the database container
  extraVolumes:
    - name: data
      emptyDir: {}
```

Resulting in a resulting README section like so:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| config.databasesToCreate[0] | string | `"postgresql"` | default database for storage of database metadata |
| config.databasesToCreate[1] | string | `"hashbash"` | database for the [hashbash](https://github.com/norwoodj/hashbash-backend-go) project |
| config.usersToCreate[0] | object | `{"admin":true,"name":"root"}` | admin user |
| config.usersToCreate[1] | object | `{"name":"hashbash","readwriteDatabases":["hashbash"]}` | user with access to the database with the same name |
| statefulset.extraVolumes | list | `[{"emptyDir":{},"name":"data"}]` | Additional volumes to be mounted into the database container |
| statefulset.image.repository | string | `"jnorwood/postgresql:11"` | Image to use for deploying, must support an entrypoint which creates users/databases from appropriate config files |
| statefulset.image.tag | string | `"18.0831"` |  |

## Installation
helm-docs can be installed using [homebrew](https://brew.sh/):

```bash
brew install blakyaks/tap/yaml-docs
```

or [scoop](https://scoop.sh):

```bash
scoop install yaml-docs
```

This will download and install the [latest release](https://github.com/blakyaks/yaml-docs/releases/latest)
of the tool.

To build from source in this repository:

```bash
cd cmd/yaml-docs
go build
```

Or install from source:

```bash
go install github.com/blakyaks/yaml-docs/cmd/yaml-docs@latest
```

## Usage

### Running the binary directly

To run and generate documentation into READMEs for all YAML files within or recursively contained by a directory:

```bash
yaml-docs --config-search-root .
# OR
yaml-docs --config-search-root . --dry-run # prints generated documentation to stdout rather than modifying READMEs
```

The tool searches recursively through subdirectories of the current directory for `.yaml` and `.yml` files and generates documentation for every file that it finds.

### Using docker

You can mount a directory with YAML files under `/yaml-docs` within the container.

Then run:

```bash
docker run --rm --volume "$(pwd):/yaml-docs" -u $(id -u) blakyaks/yaml-docs:latest
```

## Ignoring Directories

yaml-docs supports a `.yamldocsignore` file, exactly like a `.gitignore` file in which one can specify directories to ignore
when searching for YAML configuration files. Directories specified do not need to have configuration files, so parent directories containing potentially many YAML files can be ignored and none of the files underneath them will be processed. You may also directly reference the configuration file to skip processing for it.

> TODO: Additional documentation and use cases to follow.
