---
layout: default
---

**Mole** is a _cli_ application to create _ssh_ tunnels focused on resiliency
and user experience.

```sh
$ mole start local --remote :3306 --server my-database-server
INFO[0000] tunnel channel is waiting for connection      destination="127.0.0.1:3306" source="127.0.0.1:32863"
```

**Highlighted Features**

  * [Auto address selection](#let-mole-to-randomly-select-the-source-endpoint): find a port available and start listening to it, so the `--source` flag doesn't need to be given every time you run the app.
  * [Create multiple tunnels using a single ssh connection](#create-multiple-tunnels-using-a-single-ssh-connection): multiple tunnels can be established using a single connection to a ssh server by specifying different `--destination` flags.
  * [Aliases](#create-an-alias-so-there-is-no-need-to-remember-the-tunnel-settings-afterwards): save your tunnel settings under an alias, so it can be reused later.
  * Leverage the SSH Config File: use some options (e.g. user name, identity key and port), specified in *$HOME/.ssh/config* whenever possible, so there is no need to have the same SSH server configuration in multiple places.
  * Resiliency! Then tunnel will never go down if you don't want to:
    * Idle clients do not get disconnected from the ssh server since Mole keeps sending synthetic packets acting as a keep alive mechanism. 
    * Auto reconnection to the ssh server if the it is dropped by any reason.

# Table of Contents

* [Use Cases](#use-cases)
  * [Access a computer or service behind a firewall](#access-a-computer-or-service-behind-a-firewall)
  * [Access a service that is listening on a non-routable network](#access-a-service-that-is-listening-on-a-non-routable-network)
  * [Expose a service to someone outside your network](#expose-a-service-to-someone-outside-your-network)
* [Installation](#installation)
  * [Linux and Mac](#linux-and-mac)
  * [Homebrew](#or-if-you-prefer-install-it-through-homebrew)
  * [MacPorts](#or-you-can-also-install-using-macports)
  * [Windows](#windows)
* [Usage](#usage)
* [Examples](#examples)
  * [Basics](#basics)
  * [Use the ssh config file to lookup a given server host](#use-the-ssh-config-file-to-lookup-a-given-server-host)
  * [Let mole to randomly select the source endpoint](#let-mole-to-randomly-select-the-source-endpoint)
  * [Connect to a remote service that is running on 127.0.0.1 by specifying only the destination port](#connect-to-a-remote-service-that-is-running-on-127001-by-specifying-only-the-destination-port)
  * [Create an alias, so there is no need to remember the tunnel settings afterwards](#create-an-alias-so-there-is-no-need-to-remember-the-tunnel-settings-afterwards)
  * [Start mole in background](#start-mole-in-background)
  * [Leveraging LocalForward from SSH configuration file](#leveraging-localforward-from-ssh-configuration-file)
  * [Leveraging RemoteForward from SSH configuration file](#leveraging-remoteforward-from-ssh-configuration-file)
  * [Create multiple tunnels using a single ssh connection](#create-multiple-tunnels-using-a-single-ssh-connection)

# Use Cases

_...or why on Earth would I need something like this?_

## Access a computer or service behind a firewall

**Mole** can help you to **access** and/or **expose** **services** outside the perimeter network that are blocked by a firewall or unreachable, as long as the user has **ssh** access to a computer (known as **Jump Server**) with access to the target computer or service.

```ascii
+----------+          +----------+          +----------+
|          |          |          |          |          |
|          |          | Firewall |          |          |
|          |          |          |          |          |
|  Local   |  tunnel  +----------+  tunnel  |   Jump   |
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


## Access a service that is listening on a non-routable network

```ascii
+----------------------------------------------+         +--------------------+
| Local Computer                               |         | Jump Server        |
|                                              |         |                    |
|                                              | tunnel  |                    |
|                           (172.17.0.10:5001)-+---------+---> ssh server     |
|                               mole           |         | (172.17.0.100:22)  |
|                                / \           |         |      |             |
|                                 |            |         |      |             |
|                                 | forward    |         |      |             |
|   mysql \                       |            |         |      |             |
|     --host=127.0.0.1 ----> source address    |         |      | forward     |
|     --port=3306           (127.0.0.1:3306)   |         |      |             |
|     mydb                                     |         |      |             |
+----------------------------------------------+         +------+-------------+
                                                                |
                                                                |
                                                                |
                                                        +-------+-------------+
                                                        |       |    Database |
                                                        |       |      Server |
                                                        |      \ /            |
                                                        | destination address |
                                                        |  (192.168.1.1:3306) |
                                                        |                     |
                                                        +---------------------+
```

```sh
$ mole start local \
  --source 127.0.0.1:3306 \
  --destination 192.168.1.1:3306 \
  --server user@172.17.0.100
```

## Expose a service to someone outside your network

```ascii
+---------------------+         +----------------------+
|  Computer #1        |         |   Jump Server        |
|                     |         |                      |
|                     |  tunnel |                      |
|       mole  <-------+---------+-- (172.17.0.100:22)  |
|  (172.17.0.10:5001) |         |      ssh server      |
|        |            |         |       / \            |
|        | forward    |         |        |             |
|        |            |         |        | forward     |
|       \ /           |         |        |             |
|     web server      |         |   (192.168.1.1:8080) |
|  (127.0.0.1:8080)   |         |     source address   |
| destination address |         |               / \    |
|                     |         |                |     |
+---------------------+         |                |     |
                                |                |     |
                                +----------------+-----+
                                                 |
                                                 | connect
                                                 |
                                +----------------+-----+
                                | Computer #2    |     |
                                |                |     |
                                |                |     |
                                | $ curl \       |     |
                                |   192.168.1.1:8080   |
                                |                      |
                                +----------------------+
```


```sh
$ mole start remote \
  --source 192.168.1.1:8080 \
  --destination 127.0.0.1:8080 \
  --server user@172.17.0.100
```

# Installation

## Linux and Mac

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/davrodpin/mole/master/tools/install.sh)
```

### or if you prefer install it through [Homebrew](https://brew.sh)

```sh
brew tap davrodpin/homebrew-mole && brew install mole
```

### or you can also install using [MacPorts](https://www.macports.org/)

```sh
sudo port selfupdate && sudo port install mole
```

## Windows

* Download Mole for Windows from [here](https://github.com/davrodpin/mole/releases/latest)

# Usage

```
mole --help
Tool to manage ssh tunnels focused on resiliency and user experience.

Usage:
  mole [command]

Available Commands:
  add         Adds an alias for a ssh tunneling configuration
  delete      Deletes an alias for a ssh tunneling configuration
  help        Help about any command
  show        Shows configuration details about ssh tunnel aliases
  start       Starts a ssh tunnel
  stop        Stops a ssh tunnel
  version     Prints the version for mole

Flags:
  -h, --help   help for mole

Use "mole [command] --help" for more information about a command.
```  

## Examples

### Basics

```sh
$ mole start local \
    --verbose \
    --source 127.0.0.1:8080 \
    --destination 172.17.0.100:80 \
    --server user@example.com:22 \
    --key ~/.ssh/id_rsa
INFO[0000] tunnel channel is waiting for connection      destination="172.17.0.100:80" source="127.0.0.1:8080"
```

### Use the ssh config file to lookup a given server host

```sh
$ cat $HOME/.ssh/config
Host example
  User mole
  Hostname 127.0.0.1
  Port 22122
  IdentityFile test-env/ssh-server/keys/key
$ mole start local \
   --verbose \
   --source 127.0.0.1:8080 \
   --destination 172.17.0.100:80 \
   --server example
INFO[0000] tunnel channel is waiting for connection      destination="172.17.0.100:80" source="127.0.0.1:8080"
```

### Let mole to randomly select the source endpoint

```sh
$ mole start local \
    --destination 172.17.0.100:80 \
    --server example
INFO[0000] tunnel channel is waiting for connection      destination="172.17.0.100:80" source="127.0.0.1:40525"
```
### Bind the local address to 127.0.0.1 by specifying only the source port

```sh
$ mole start local \
    --source :8080 \
    --destination 172.17.0.100:80 \
    --server example
INFO[0000] tunnel channel is waiting for connection      destination="172.17.0.100:80" source="127.0.0.1:8080"
```

### Connect to a remote service that is running on 127.0.0.1 by specifying only the destination port

```sh
$ mole start local \
    --source 127.0.0.1:8080 \
    --destination :80 \
    --server example
INFO[0000] tunnel channel is waiting for connection      destination="127.0.0.1:80" source="127.0.0.1:8080"
```

### Create an alias, so there is no need to remember the tunnel settings afterwards

```sh
$ mole add alias local example \
    --source :8080 \
    --destination 172.17.0.100:80 \
    --server user@example.com:22
$ mole start alias example
INFO[0000] tunnel channel is waiting for connection      destination="172.17.0.100:80" source="127.0.0.1:8080"
```

### Start mole in background

```sh
$ mole add alias local example \
    --source :8080 \
    --destination 172.17.0.100:80 \
    --server user@example.com:22 \
    --detach
$ mole start alias example
INFO[0000] execute "mole stop example" if you like to stop it at any time
```

### Leveraging LocalForward from SSH configuration file

```sh
$ cat ~/.ssh/config
Host example
  User mole
  Hostname 127.0.0.1
  Port 22122
  LocalForward 21112 192.168.33.11:80
  IdentityFile test-env/ssh-server/keys/key
$ mole start local --server example 
INFO[0000] tunnel channel is waiting for connection      destination="192.168.33.11:80" source="127.0.0.1:21112"
```

### Leveraging RemoteForward from SSH configuration file

```sh
$ cat ~/.ssh/config
Host example
  User mole
  Hostname 127.0.0.1
  Port 22122
  LocalForward 192.168.33.11:9090 172.0.0.1:8080
  IdentityFile test-env/ssh-server/keys/key
$ mole start remote --server example
INFO[0000] tunnel channel is waiting for connection      destination="192.168.33.11:8080" source="127.0.0.1:9090"
```

### Create multiple tunnels using a single ssh connection

```sh
$ mole start local \
    --source :9090 \
    --source :9091 \
    --destination 192.168.33.11:80 \
    --destination 192.168.33.11:8080 \
    --server example
INFO[0000] tunnel channel is waiting for connection      destination="192.168.33.11:8080" source="127.0.0.1:9091"
INFO[0000] tunnel channel is waiting for connection      destination="192.168.33.11:80" source="127.0.0.1:9090"
```

