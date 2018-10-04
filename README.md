[![Documentation](https://godoc.org/github.com/davrodpin/mole?status.svg)](http://godoc.org/github.com/davrodpin/mole)
[![Build Status](https://travis-ci.org/davrodpin/mole.svg?branch=master)](https://travis-ci.org/davrodpin/mole)
[![Go Report Card](https://goreportcard.com/badge/github.com/davrodpin/mole)](https://goreportcard.com/report/github.com/davrodpin/mole)
[![codebeat badge](https://codebeat.co/badges/ec5e4267-3292-4ef4-818c-b58e94a5dbbb)](https://codebeat.co/projects/github-com-davrodpin-mole-master)
[![codecov](https://codecov.io/gh/davrodpin/mole/branch/master/graph/badge.svg)](https://codecov.io/gh/davrodpin/mole)
# Mole 

Mole is a cli application to create ssh tunnels, forwarding a local port to a
remote endpoint through an ssh server.

It tries to leverage the user's ssh config file, *$HOME/.ssh/config*, to find
options to be used to connect to the ssh server when those options are not
given (e.g. user name, identity key path and port).

## How to install

```sh
$ go get -d github.com/davrodpin/mole/... && go install github.com/davrodpin/mole/cmd/mole
```

## Examples

```sh
$ mole -local 127.0.0.1:8080 -remote 172.17.0.100:80 -server user@example.com:22 -i ~/.ssh/id_rsa
```

### 

```sh
$ cat $HOME/.ssh/config
Host example1
  Hostname 10.0.0.12
  Port 2222
  Username user
  IdentityFile ~/.ssh/id_rsa
$ mole -local 127.0.0.1:8080 -remote 172.17.0.100:80 -server example1 
```
