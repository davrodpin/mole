[![CI](https://github.com/davrodpin/mole/actions/workflows/ci-master.yml/badge.svg)](https://github.com/davrodpin/mole/actions/workflows/ci-master.yml)
[![Documentation](https://godoc.org/github.com/davrodpin/mole?status.svg)](http://godoc.org/github.com/davrodpin/mole)
# [Mole](https://davrodpin.github.io/mole/)

Mole is a cli application to create ssh tunnels focused on resiliency and user
experience.

For more information about installation, usage, examples and specific use cases,
please visit https://davrodpin.github.io/mole/

## How to build from source

[Go 1.17.1+](https://golang.org/dl/) is required to be installed on your system to
build this project.

```sh
$ make build
```

## How to run tests

```sh
$ make test
```

## How to generate a code coverage report

```sh
$ make cover && open coverage.html
```

## How to run static analysis

1. Install [golangci-lint](https://golangci-lint.run/usage/install/)

2. Run the following command

```sh
$ make lint
```

# Test Environment

The project provides a small automated infrastructure to help with manual testing
Please refer to [this document](test-env/README.md) for more details about it.

# How to Contribute

Please refere to [CONTRIBUTING.md](https://github.com/davrodpin/mole/blob/master/CONTRIBUTING.md)
for details on how to contribute to this project.
