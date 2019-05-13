[![Documentation](https://godoc.org/github.com/davrodpin/mole?status.svg)](http://godoc.org/github.com/davrodpin/mole)
# [Mole](https://davrodpin.github.io/mole/)

Mole is a cli application to create ssh tunnels, forwarding a local port to a
remote endpoint through an ssh server.

For more information about usage, examples and specific use cases, please visit https://davrodpin.github.io/mole/

## How to build

### Build and Install from Source

* [Go 1.12.5+](https://golang.org/dl/)

```sh
$ go install github.com/davrodpin/mole/cmd/mole
```

# Test Environment

The project provides a small automated infrastructure to help on funcional
tests. Please refer to [this document](test-env/README.md) for more details about it.
