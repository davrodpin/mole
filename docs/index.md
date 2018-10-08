---
layout: default
---

**Mole** is a _cli_ application to create _ssh_ tunnels, forwarding a local port to a remote address through a _ssh_ server.

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
  -server example@172.12.0.100
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


## Linux

```sh
curl -L https://github.com/davrodpin/mole/releases/download/v0.0.1/mole0.0.1.linux-amd64.tar.gz | tar xz -C /usr/local/bin
```

## macOS

```sh
curl -L https://github.com/davrodpin/mole/releases/download/v0.0.1/mole0.0.1.darwin-amd64.tar.gz | tar xz -C /usr/local/bin
```

# Usage

```sh
$ mole -help
usage:
  mole [-v] [-local <host>:<port>] -remote <host>:<port> -server [<user>@]<host>[:<port>] [-key <key_path>]
  mole -help

  -help
        list all options available
  -key string
        server authentication key file path
  -local value
        local endpoint address: <host>:<port>
  -remote value
        remote endpoing address: <host>:<port>
  -server value
        server address: [<user>@]<host>[:<port>]
  -v    increases log verbosity
```  
