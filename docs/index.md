---
layout: default
---

**Mole** is a _cli_ application to create _ssh_ tunnels, forwarding a local port to a remote address through a _ssh_ server.

```sh
$ mole -remote :3306 -server my-database-server
INFO[0000] listening on local address                    local_address="127.0.0.1:51082"
```

**Highlighted Features**

  * [Auto local address selection](#let-mole-to-randomly-select-the-local-endpoint): find a port available and start listening to it, so the `-local` flag doesn't need to be given every time you run the app.
  * [Create multiple tunnels using a single ssh connection](#create-multiple-tunnels-using-a-single-ssh-connection): multiple tunnels can be established using a single connection to a ssh server by specifying different `-remote` flags.
  * [Aliases](#create-an-alias-so-there-is-no-need-to-remember-the-tunnel-settings-afterwards): save your tunnel settings under an alias, so it can be reused later.
  * Leverage the SSH Config File: use some options (e.g. user name, identity key and port), specified in *$HOME/.ssh/config* whenever possible, so there is no need to have the same SSH server configuration in multiple places.
  * Resiliency! Then tunnel will never go down if you don't want it to:
    * Idle clients do not get disconnected from the ssh server since Mole keeps sending synthetic packets acting as a [keep alive mechanism](#configure-the-time-interval-to-send-keep-alive-packets). 
    * Auto [reconnection to the ssh server]() if the it is dropped by any reason.

# Table of Contents

* [Use Cases](#use-cases)
  * [Access a computer or service behind a firewall](#access-a-computer-or-service-behind-a-firewall)
  * [Access a service that is listening only on a local address](#access-a-service-that-is-listening-only-on-a-local-address)
* [Installation](#installation)
  * [Linux and Mac](#linux-and-mac)
  * [Homebrew](#or-if-you-prefer-install-it-through-homebrew)
  * [Windows](#windows)
* [Usage](#usage)
* [Examples](#examples)
  * [Basics](#basics)
  * [Use the ssh config file to lookup a given server host](#use-the-ssh-config-file-to-lookup-a-given-server-host)
  * [Let mole to randomly select the local endpoint](#let-mole-to-randomly-select-the-local-endpoint)
  * [Connect to a remote service that is running on 127.0.0.1 by specifying only the remote port](#connect-to-a-remote-service-that-is-running-on-127001-by-specifying-only-the-remote-port)
  * [Create an alias, so there is no need to remember the tunnel settings afterwards](#create-an-alias-so-there-is-no-need-to-remember-the-tunnel-settings-afterwards)
  * [Start mole in background](#start-mole-in-background)
  * [Leveraging LocalForward from SSH configuration file](#leveraging-localforward-from-ssh-configuration-file)
  * [Create multiple tunnels using a single ssh connection](#create-multiple-tunnels-using-a-single-ssh-connection)
  * [Configure the time interval to send keep alive packets](#configure-the-time-interval-to-send-keep-alive-packets)
  * [Configure connection retries and retry wait interval](#configure-connection-retries-and-retry-wait-interval)

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

## Linux and Mac

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/davrodpin/mole/master/tools/install.sh)
```

### or if you prefer install it through [Homebrew](https://brew.sh)

```sh
brew tap davrodpin/homebrew-mole && brew install mole
```

## Windows

* Download Mole for Windows from [here](https://github.com/davrodpin/mole/releases/latest)

# Usage

```
$ ./mole -help
usage:
        mole [-v] [-insecure] [-detach] (-local [<host>]:<port>)... (-remote [<host>]:<port>)... -server [<user>@]<host>[:<port>] [-key <key_path>] [-keep-alive-interval <time_interval>] [-connection-retries <retries>] [-retry-wait <time>]
        mole -alias <alias_name> [-v] (-local [<host>]:<port>)... (-remote [<host>]:<port>)... -server [<user>@]<host>[:<port>] [-key <key_path>] [-keep-alive-interval <time_interval>] [-connection-retries <retries>] [-retry-wait <time>]
        mole -alias <alias_name> -delete
        mole -start <alias_name>
        mole -help
        mole -version

  -alias string
        create a tunnel alias
  -aliases
        list all aliases
  -connection-retries int
        (optional) maximum number of connection retries to the ssh server. Provide 0 if mole should never give up or negative number to disable retries (default 3)
  -delete
        delete a tunnel alias (must be used with -alias)
  -detach
        (optional) run process in background
  -help
        list all options available
  -insecure
        (optional) skip host key validation when connecting to ssh server
  -keep-alive-interval duration
        (optional) time interval for keep alive packets to be sent (default 10s)
  -key string
        (optional) Set server authentication key file path
  -local value
        (optional) set local endpoint address: [<host>]:<port>. Multiple -local args can be provided
  -remote value
        (optional) set remote endpoint address: [<host>]:<port>. Multiple -remote args can be provided
  -retry-wait duration
        (optional) time to wait before trying to reconnect to ssh server (default 3s)
  -server value
        set server address: [<user>@]<host>[:<port>]
  -start string
        start a tunnel using a given alias
  -stop string
        stop background process
  -timeout duration
        (optional) ssh server connection timeout (default 3s)
  -v    (optional) Increase log verbosity
  -version
        display the mole version
```  

## Examples

### Basics

```sh
$ mole -v -local 127.0.0.1:8080 -remote 172.17.0.100:80 -server user@example.com:22 -key ~/.ssh/id_rsa
$ ./mole -v -local 127.0.0.1:8080 -remote 172.17.0.100:80 -server user@example.com:22 -key ~/.ssh/id_rsa
DEBU[0000] cli options                                   options="[local=127.0.0.1:8080, remote=172.17.0.100:80, server=user@example.com:22, key=/Users/mole/.ssh/id_rsa, verbose=true, help=false, version=false, detach=false]"
DEBU[0000] using ssh config file from: /Users/mole/.ssh/config
DEBU[0000] server: [name=example.com, address=example.com:22, user=user]
DEBU[0000] tunnel: [channels:[[local=127.0.0.1:8080, remote=172.17.0.100:80]], server:example.com:22]
DEBU[0000] known_hosts file used: /Users/mole/.ssh/known_hosts
INFO[0000] tunnel is ready                               local="127.0.0.1:8080" remote="172.17.0.100:80"
```

### Use the ssh config file to lookup a given server host

```sh
$ cat $HOME/.ssh/config
Host example1
  User mole
  Hostname 127.0.0.1
  Port 22122
  IdentityFile test-env/ssh-server/keys/key
$ mole -v -local 127.0.0.1:8080 -remote 192.168.33.11:80 -server example1
DEBU[0000] cli options                                   options="[local=127.0.0.1:8080, remote=192.168.33.11:80, server=example1, key=, verbose=true, help=false, version=false, detach=false]"
DEBU[0000] using ssh config file from: /Users/mole/.ssh/config
DEBU[0000] server: [name=example1, address=127.0.0.1:22122, user=mole]
DEBU[0000] tunnel: [channels:[[local=127.0.0.1:8080, remote=192.168.33.11:80]], server:127.0.0.1:22122]
DEBU[0000] known_hosts file used: /Users/mole/.ssh/known_hosts
DEBU[0000] new connection established to server          server="[name=example1, address=127.0.0.1:22122, user=mole]"
INFO[0000] tunnel is ready                               local="127.0.0.1:8080" remote="192.168.33.11:80"
```

### Let mole to randomly select the local endpoint

```sh
$ mole -remote 172.17.0.100:80 -server example1
INFO[0000] tunnel is ready                               local="127.0.0.1:61305" remote="192.168.33.11:80"
```
### Bind the local address to 127.0.0.1 by specifying only the local port

```sh
$ mole -v -local :8080 -remote 192.168.33.10:80 -server example1
DEBU[0000] cli options                                   options="[local=127.0.0.1:8080, remote=192.168.33.11:80, server=example1, key=, verbose=true, help=false, version=false, detach=false]"
DEBU[0000] using ssh config file from: /Users/mole/.ssh/config
DEBU[0000] server: [name=example1, address=127.0.0.1:22122, user=mole]
DEBU[0000] tunnel: [channels:[[local=127.0.0.1:8080, remote=192.168.33.11:80]], server:127.0.0.1:22122]
DEBU[0000] known_hosts file used: /Users/mole/.ssh/known_hosts
DEBU[0000] new connection established to server          server="[name=example1, address=127.0.0.1:22122, user=mole]"
INFO[0000] tunnel is ready                               local="127.0.0.1:8080" remote="192.168.33.11:80"
```

### Connect to a remote service that is running on 127.0.0.1 by specifying only the remote port

```sh
$ mole -v -local 127.0.0.1:8080 -remote :80 -server example2
DEBU[0000] cli options                                   options="[local=127.0.0.1:8080, remote=127.0.0.1:80, server=example1, key=, verbose=true, help=false, version=false, detach=false]"
DEBU[0000] using ssh config file from: /Users/mole/.ssh/config
DEBU[0000] server: [name=example1, address=127.0.0.1:22222, user=mole]
DEBU[0000] tunnel: [channels:[[local=127.0.0.1:8080, remote=127.0.0.1:80]], server:127.0.0.1:22222]
DEBU[0000] known_hosts file used: /Users/mole/.ssh/known_hosts
DEBU[0000] new connection established to server          server="[name=example1, address=127.0.0.1:22222, user=mole]"
INFO[0000] tunnel is ready                               local="127.0.0.1:8080" remote="127.0.0.1:80"
```

### Create an alias, so there is no need to remember the tunnel settings afterwards

```sh
$ mole -alias example1 -v -local :8443 -remote :443 -server mole@example.com
$ mole -start example1
DEBU[0000] cli options                                   options="[local=127.0.0.1:8443, remote=127.0.0.1:443, server=mole@example.com, key=, verbose=true, help=false, version=false, detach=false]"
DEBU[0000] using ssh config file from: /Users/mole/.ssh/config
DEBU[0000] server: [name=example.com, address=127.0.0.1:22222, user=mole]
DEBU[0000] tunnel: [channels:[[local=127.0.0.1:8443, remote=127.0.0.1:443]], server:127.0.0.1:22222]
DEBU[0000] known_hosts file used: /Users/mole/.ssh/known_hosts
DEBU[0000] new connection established to server          server="[name=example.com, address=127.0.0.1:22222, user=mole]"
INFO[0000] tunnel is ready                               local="127.0.0.1:8443" remote="127.0.0.1:443"

```

### Start mole in background

```sh
$ mole -alias example2 -v -local :8443 -remote :443 -server user@example.com
$ mole -start example2 -detach
INFO[0000] execute "mole -stop example2" if you like to stop it at any time
$ tail -f ~/.mole/instances/example2/mole.log
time="2019-05-13T09:56:57-07:00" level=info msg="listening on local address" local_address="127.0.0.1:21112"
$ mole -stop example2
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
$ mole -v -server example1
DEBU[0000] cli options                                   options="[local=, remote=, server=example1, key=, verbose=true, help=false, version=false, detach=false]"
DEBU[0000] using ssh config file from: /Users/mole/.ssh/config
DEBU[0000] server: [name=example1, address=127.0.0.1:22122, user=mole]
DEBU[0000] using ssh config file from: /Users/mole/.ssh/config
DEBU[0000] tunnel: [channels:[[local=127.0.0.1:21112, remote=192.168.33.10:80]], server:127.0.0.1:22122]
DEBU[0000] known_hosts file used: /Users/mole/.ssh/known_hosts
DEBU[0000] new connection established to server          server="[name=example1, address=127.0.0.1:22122, user=mole]"
INFO[0000] tunnel is ready                               local="127.0.0.1:21112" remote="192.168.33.10:80"
```

### Create multiple tunnels using a single ssh connection

```sh
$ mole -v -local :8080 -local :3306 -remote 172.17.0.1:80 -remote 172.17.0.2:3306 -server example1
DEBU[0000] cli options                                   options="[local=:8080,:3306, remote=172.17.0.1:80,172.17.0.2:3306, server=example1, key=, verbose=true, help=false, version=false, detach=false]"
DEBU[0000] using ssh config file from: /Users/mole/.ssh/config
DEBU[0000] server: [name=example1, address=127.0.0.1:22122, user=mole]
DEBU[0000] tunnel: [channels:[[local=127.0.0.1:8080, remote=172.17.0.1:80] [local=127.0.0.1:3306, remote=172.17.0.2:3306]], server:127.0.0.1:22122]
DEBU[0000] known_hosts file used: /Users/mole/.ssh/known_hosts
DEBU[0000] new connection established to server          server="[name=example1, address=127.0.0.1:22122, user=mole]"
INFO[0000] tunnel is ready                               local="127.0.0.1:3306" remote="172.17.0.2:3306"
INFO[0000] tunnel is ready                               local="127.0.0.1:8080" remote="172.17.0.1:80"
```

### Configure the time interval to send keep alive packets

```sh
$ ./mole -v -local :8080 -remote 172.17.0.1:80 -server example1 -keep-alive-interval 2s
DEBU[0000] cli options                                   options="[local=:8080, remote=172.17.0.1:80, server=example1, key=, verbose=true, help=false, version=false, detach=false, insecure=false, keep-alive-interval=2s, timeout=3s, connection-retries=3, retry-wait=3s]"
DEBU[0000] using ssh config file from: /Users/mole/.ssh/config
DEBU[0000] server: [name=example1, address=127.0.0.1:22122, user=mole]
DEBU[0000] tunnel: [channels:[[local=127.0.0.1:8080, remote=172.17.0.1:80]], server:127.0.0.1:22122]
DEBU[0000] known_hosts file used: /Users/mole/.ssh/known_hosts
DEBU[0000] connection to the ssh server is established   server="[name=example1, address=127.0.0.1:22122, user=mole]"
DEBU[0000] start sending keep alive packets
INFO[0000] tunnel channel is waiting for connection      local="127.0.0.1:8080" remote="172.17.0.1:80"
```

### Configure connection retries and retry wait interval

```sh
$ ./mole -v -local :8080 -remote 172.17.0.1:80 -server example1 --connection-retries 5 --retry-wait 10s
DEBU[0000] cli options                                   options="[local=:8080, remote=172.17.0.1:80, server=example1, key=, verbose=true, help=false, version=false, detach=false, insecure=false, keep-alive-interval=2s, timeout=3s, connection-retries=5, retry-wait=10s]"
DEBU[0000] using ssh config file from: /Users/mole/.ssh/config
DEBU[0000] server: [name=example1, address=127.0.0.1:22122, user=mole]
DEBU[0000] tunnel: [channels:[[local=127.0.0.1:8080, remote=172.17.0.1:80]], server:127.0.0.1:22122]
DEBU[0000] known_hosts file used: /Users/mole/.ssh/known_hosts
DEBU[0000] connection to the ssh server is established   server="[name=example1, address=127.0.0.1:22122, user=mole]"
DEBU[0000] start sending keep alive packets
INFO[0000] tunnel channel is waiting for connection      local="127.0.0.1:8080" remote="172.17.0.1:80"
```

