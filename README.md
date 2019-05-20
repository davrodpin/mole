[![Documentation](https://godoc.org/github.com/davrodpin/mole?status.svg)](http://godoc.org/github.com/davrodpin/mole)
# [Mole](https://davrodpin.github.io/mole/)

Mole is a cli application to create ssh tunnels, forwarding a local port to a
remote endpoint through a ssh server.

For more information about usage, examples and specific use cases, please visit https://davrodpin.github.io/mole/

## How to build

### Build and Install from Source

* [Go 1.12.5+](https://golang.org/dl/) is required
* Mole uses [Go Modules](https://blog.golang.org/using-go-modules) to manage its dependencies, so remember to clone the project outside `GOPATH` and unset it.

```sh
$ go build github.com/davrodpin/mole/cmd/mole
```

# Test Environment

The project provides a small automated infrastructure to help on funcional
tests. Please refer to [this document](test-env/README.md) for more details about it.
