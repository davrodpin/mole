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

### Provide all supported options

```sh
$ mole -v -local 127.0.0.1:8080 -remote 172.17.0.100:80 -server user@example.com:22 -key ~/.ssh/id_rsa
DEBU[0000] cli options                                   key=/home/mole/.ssh/id_rsa local="127.0.0.1:8080" remote="172.17.0.100:80" server="user@example.com:22" v=true
DEBU[0000] using ssh config file from: /home/mole/.ssh/config
DEBU[0000] server: [name=example.com, address=example.com:22, user=user, key=/home/mole/.ssh/id_rsa]
DEBU[0000] tunnel: [local:127.0.0.1:8080, server:example.com:22, remote:172.17.0.100:80]
INFO[0000] listening on local address                    local_address="127.0.0.1:8080"
```

### Use the ssh config file to lookup a given server host

```sh
$ cat $HOME/.ssh/config
Host example1
  Hostname 10.0.0.12
  Port 2222
  User user
  IdentityFile ~/.ssh/id_rsa
$ mole -v -local 127.0.0.1:8080 -remote 172.17.0.100:80 -server example1
DEBU[0000] cli options                                   key= local="127.0.0.1:8080" remote="172.17.0.100:80" server=example1 v=true
DEBU[0000] using ssh config file from: /home/mole/.ssh/config
DEBU[0000] server: [name=example1, address=10.0.0.12:2222, user=user, key=/home/mole/.ssh/id_rsa]
DEBU[0000] tunnel: [local:127.0.0.1:8080, server:10.0.0.12:2222, remote:172.17.0.100:80]
INFO[0000] listening on local address                    local_address="127.0.0.1:8080"
```

### Let mole to randomly select the local endpoint

```sh
$ mole -remote 172.17.0.100:80 -server example1 
INFO[0000] listening on local address                    local_address="127.0.0.1:61305"
```
### Bind the local address to 127.0.0.1 by specifying only the local port

```sh
$ mole -v -local :8080 -remote 172.17.0.100:80 -server example1
DEBU[0000] cli options                                   key= local="127.0.0.1:8080" remote="172.17.0.100:80" server=example1 v=true
DEBU[0000] using ssh config file from: /home/mole/.ssh/config
DEBU[0000] server: [name=example1, address=10.0.0.12:2222, user=user, key=/home/mole/.ssh/id_rsa]
DEBU[0000] tunnel: [local:127.0.0.1:8080, server:10.0.0.12:2222, remote:172.17.0.100:80]
INFO[0000] listening on local address                    local_address="127.0.0.1:8080"
```

### Connect to a remote ip address 127.0.0.1 by specifying only the remote port

```sh
$ mole -v -local 127.0.0.1:8080 -remote :80 -server example1
DEBU[0000] cli options                                   key= local="127.0.0.1:8080" remote="127.0.0.1:80" server=example1 v=true
DEBU[0000] using ssh config file from: /home/mole/.ssh/config
DEBU[0000] server: [name=example1, address=10.0.0.12:2222, user=user, key=/home/mole/.ssh/id_rsa]
DEBU[0000] tunnel: [local:127.0.0.1:8080, server:10.0.0.12:2222, remote:127.0.0.1:80]
INFO[0000] listening on local address                    local_address="127.0.0.1:8080"
```
