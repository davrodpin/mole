[![Documentation](https://godoc.org/github.com/davrodpin/mole?status.svg)](http://godoc.org/github.com/davrodpin/mole)
[![Build Status](https://travis-ci.org/davrodpin/mole.svg?branch=master)](https://travis-ci.org/davrodpin/mole)
[![Go Report Card](https://goreportcard.com/badge/github.com/davrodpin/mole)](https://goreportcard.com/report/github.com/davrodpin/mole)
[![codebeat badge](https://codebeat.co/badges/ec5e4267-3292-4ef4-818c-b58e94a5dbbb)](https://codebeat.co/projects/github-com-davrodpin-mole-master)
[![codecov](https://codecov.io/gh/davrodpin/mole/branch/master/graph/badge.svg)](https://codecov.io/gh/davrodpin/mole)
# [Mole](https://davrodpin.github.io/mole/)

Mole is a cli application to create ssh tunnels, forwarding a local port to a
remote endpoint through an ssh server.

For more information about usage, examples and specific use cases, please visit https://davrodpin.github.io/mole/

## How to install

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/davrodpin/mole/master/tools/install.sh)
```

### or if you prefer install it through [Homebrew](https://brew.sh)

```sh
brew tap davrodpin/homebrew-mole && brew install mole
```

## How to use

```sh
$ mole -V --remote :443 --server user@example.com
DEBU[0000] cli options                                   options="[local=, remote=:443, server=user@example.com, key=, verbose=true, help=false, version=false]"
DEBU[0000] using ssh config file from: /home/mole/.ssh/config
DEBU[0000] server: [name=example.com, address=example.com:22, user=user, key=/home/mole/.ssh/id_rsa]
DEBU[0000] tunnel: [local:127.0.0.1:63046, server:example.com:22, remote:127.0.0.1:443]
INFO[0000] listening on local address                    local_address="127.0.0.1:63046"
```

```sh
$ mole -V --local 127.0.0.1:8080 --remote 172.17.0.100:80 --server user@example.com:22 --key ~/.ssh/id_rsa
DEBU[0000] cli options                                   key=/home/mole/.ssh/id_rsa local="127.0.0.1:8080" remote="172.17.0.100:80" server="user@example.com:22" v=true
DEBU[0000] using ssh config file from: /home/mole/.ssh/config
DEBU[0000] server: [name=example.com, address=example.com:22, user=user, key=/home/mole/.ssh/id_rsa]
DEBU[0000] tunnel: [local:127.0.0.1:8080, server:example.com:22, remote:172.17.0.100:80]
INFO[0000] listening on local address                    local_address="127.0.0.1:8080"
```

```sh
$ mole --alias example1 -V --local :8443 --remote :443 --server user@example.com
$ mole --start example1
DEBU[0000] cli options                                   options="[local=:8443, remote=:443, server=user@example.com, key=, verbose=true, help=false, version=false]"
DEBU[0000] using ssh config file from: /home/mole/.ssh/config
DEBU[0000] server: [name=example.com, address=example.com:22, user=user, key=/home/mole/.ssh/id_rsa]
DEBU[0000] tunnel: [local:127.0.0.1:8443, server:example.com:22, remote:127.0.0.1:443]
INFO[0000] listening on local address                    local_address="127.0.0.1:8443"
```

# Commands
```sh
  -a, --alias string
        Create a tunnel alias
  -A  --aliases
        List all aliases
  -d, --delete
        Delete a tunnel alias (must be used with -alias)
  -h, --help
        List all options available
  -k, --key string
        (optional) Set server authentication key file path
  -l, --local value
        (optional) Set local endpoint address: [<host>]:<port>
  -r, --remote value
        Set remote endpoint address: [<host>]:<port>
  -s, --server value
        Set server address: [<user>@]<host>[:<port>]
 -st, --start string
        Start a tunnel using a given alias (optional) Increase log verbosity
  -v  --version
        Display the mole version	
  -V  --verbose
        (optional) Increase log verbosity
```
