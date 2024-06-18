# Remote Standby

This service operates as a backup controller for [remote-commands-handler](github.com/EvergenEnergy/remote-commands-handler). If it detects a lack of commands arriving for the handler to process, it will step in and issue commands in order to maintain a basic level of control.

[![Coverage Status](https://coveralls.io/repos/github/EvergenEnergy/remote-standby/badge.svg?branch=main)](https://coveralls.io/github/EvergenEnergy/remote-standby?branch=main)

## Getting Started

All dependencies can be run using docker compose. To run locally, create your own `.env` file.

To run from the command line:

```sh
docker compose up -d
set -o allexport; source .env; set +o allexport
go run .
```

### Prerequisites

* Go v1.22+
* golangci-lint v1.55+

## Running tests

The flag `-short` will skip integration tests which require running Docker.

```sh
go test -race -short -v ./...
```

It is however advised to use `make`:

```sh
make test.unit
```

The following will run integration tests only:

```sh
make test.integration
```

## Linting

This project uses golangci-lint for checking coding style.

Install the latest version of golangci-lint.

```sh
golangci-lint run
```
