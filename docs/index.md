---
layout: default
---

**Mole** is a _cli_ application to create _ssh_ tunnels, forwarding a local port to a remote address through a _ssh_ server.


**Highlighted Features**

  * [Auto local address selection](#let-mole-to-randomly-select-the-local-endpoint): find a port available and start linstening to it, so the `-local` flag doesn't need to be given every time you run the app.
  * [Aliases](#create-an-alias-so-there-is-no-need-to-remember-the-tunnel-settings-afterwards): save your tunnel settings under an alias, so it can be reused later.
  * Leverage the SSH Config File: use some options (e.g. user name, identity key and port), specified in *$HOME/.ssh/config* whenever possible, so there is no need to have the same SSH server configuration in multiple places.

# Table of Contents

* [Use Cases](#use-cases)
  * [Access a computer or service behind a firewall](#access-a-computer-or-service-behind-a-firewall)
  * [Access a service that is listening only on a local address](#access-a-service-that-is-listening-only-on-a-local-address)
* [Installation](#installation)
  * [macOS](#macOS)
  * [Linux](#linux)
* [Usage](#usage)
* [Examples](#examples)
  * [Provide all supported options](#provide-all-supported-options)
  * [Use the ssh config file to lookup a given server host](#use-the-ssh-config-file-to-lookup-a-given-server-host)
  * [Let mole to randomly select the local endpoint](#let-mole-to-randomly-select-the-local-endpoint)
  * [Connect to a remote service that is running on 127.0.0.1 by specifying only the remote port](#connect-to-a-remote-service-that-is-running-on-127001-by-specifying-only-the-remote-port)
  * [Create an alias, so there is no need to remember the tunnel settings afterwards](#create-an-alias-so-there-is-no-need-to-remember-the-tunnel-settings-afterwards)

# Use Cases

_...or why on Earth would I need something like this?_

## Access a computer or service behind a firewall

**Mole** can help you to access computers and services outside the perimeter network that are blocked by a firewall, as long as the user has _ssh_ access to a computer with access to the target computer or service.

```ascii
+----------+          +----------+          +----------+
|          |          |          |          |          |
|          |          | Firewall |          |          |
|          |          |          |          |          |
|  Local   |  tunnel  +----------+  tunnel  |          |
| Computer |--------------------------------|  Server  |
|          |          +----------+          |          |
|          |          |          |          |          |
|          |          | Firewall |          |          |
|          |          |          |          |          |
+----------+          +----------+          +----------+
                                                 |
                                                 |
                                                 | tunnel
                                                 |
                                                 |
                                            +----------+
                                            |          |
                                            |          |
                                            |          |
                                            |          |
                                            |  Remote  |
                                            | Computer |
                                            |          |
                                            |          |
                                            |          |
                                            +----------+
```

NOTE: _Server and Remote Computer could potentially be the same machine._

## Access a service that is listening only on a local address

```sh
$ mole \
  -local 127.0.0.1:3306 \
  -remote 127.0.0.1:3306 \
  -server example@172.17.0.100
```

```ascii
+-------------------+             +--------------------+
| Local Computer    |             | Remote / Server    |
|                   |             |                    |
|                   |             |                    |
| (172.17.0.10:     |    tunnel   |                    |
|        50001)     |-------------| (172.17.0.100:22)  |
|  tunnel client    |             |  tunnel server     |
|       |           |             |         |          |
|       | port      |             |         | port     |
|       | forward   |             |         | forward  |
|       |           |             |         |          |
| (127.0.0.1:3306)  |             | (127.0.0.1:50000)  |
|  local address    |             |         |          |
|                   |             |         | local    |
|                   |             |         | conn.    |
|                   |             |         |          |
|                   |             | (127.0.0.1:3306)   |
|                   |             |  remote address    |
|                   |             |      +----+        |
|                   |             |      | DB |        |
|                   |             |      +----+        |
+-------------------+             +--------------------+
```

NOTE: _Server and Remote Computer could potentially be the same machine._

# Installation

## macOS

```sh
brew tap davrodpin/homebrew-mole && brew install mole
```

## Linux

```sh
curl -L https://github.com/davrodpin/mole/releases/download/v0.2.0/mole0.2.0.linux-amd64.tar.gz | tar xz -C /usr/local/bin
```

# Usage

```sh
$ mole -help
usage:
  mole [-v] [-local [<host>]:<port>] -remote [<host>]:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
  mole -alias <alias_name> [-v] [-local [<host>]:<port>] -remote [<host>]:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
  mole -alias <alias_name> -delete
  mole -start <alias_name>
  mole -help
  mole -version

  -alias string
        Create a tunnel alias
  -delete
        delete a tunnel alias (must be used with -alias)
  -help
        list all options available
  -key string
        (optional) Set server authentication key file path
  -local value
        (optional) Set local endpoint address: [<host>]:<port>
  -remote value
        set remote endpoing address: [<host>]:<port>
  -server value
        set server address: [<user>@]<host>[:<port>]
  -start string
        Start a tunnel using a given alias
  -v    (optional) Increase log verbosity
  -version
        display the mole version
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

### Connect to a remote service that is running on 127.0.0.1 by specifying only the remote port

```sh
$ mole -v -local 127.0.0.1:8080 -remote :80 -server example1
DEBU[0000] cli options                                   key= local="127.0.0.1:8080" remote="127.0.0.1:80" server=example1 v=true
DEBU[0000] using ssh config file from: /home/mole/.ssh/config
DEBU[0000] server: [name=example1, address=10.0.0.12:2222, user=user, key=/home/mole/.ssh/id_rsa]
DEBU[0000] tunnel: [local:127.0.0.1:8080, server:10.0.0.12:2222, remote:127.0.0.1:80]
INFO[0000] listening on local address                    local_address="127.0.0.1:8080"
```

### Create an alias, so there is no need to remember the tunnel settings afterwards

```sh
$ mole -alias example1 -v -local :8443 -remote :443 -server user@example.com
$ mole -start example1
DEBU[0000] cli options                                   options="[local=:8443, remote=:443, server=user@example.com, key=, verbose=true, help=false, version=false]"
DEBU[0000] using ssh config file from: /home/mole/.ssh/config
DEBU[0000] server: [name=example.com, address=example.com:22, user=user, key=/home/mole/.ssh/id_rsa]
DEBU[0000] tunnel: [local:127.0.0.1:8443, server:example.com:22, remote:127.0.0.1:443]
INFO[0000] listening on local address                    local_address="127.0.0.1:8443"
```
