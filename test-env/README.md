# Mole's Test Environment

This provides a small envorinment where `mole` functions can be tested and
debugged.

Once created, the test environment will provide a container running a ssh
server and another container running an http server, so ssh tunnels can be
created using `mole`.

In addition to that, the test environment provides the infrastructure to
analyze the traffic going through the ssh traffic.

## Topology

```ascii
+-------------------+                                 
|                   |
|                   |
|  Local Computer   |
|                   |
|                   |
|                   |
+-----------+-------+                                 
            | 127.0.0.1:22122
            |
            |
            |
            | 
+-----------+-------+          +---------------------+
|                   |          |                     |
|  mole_ssh         |          |  mole_http          |
|  SSH Server (:22) |          |  HTTP Server (:80)  |
|  (192.168.33.10)  |----------|  (192.168.33.11)    |
|                   |          |                     |
|                   |          |                     |
+-------------------+          +---------------------+
```

## Required Software

You will need the following software installed on your computer to build this
test environment:

* [Docker](https://docs.docker.com/install/)
* ssh
* ssh-keygen
* [Wireshark](https://www.wireshark.org/download.html) (optional)

## Managing the Environment

### Setup

```sh
$ make test-env
```

This builds two docker containers: `mole_ssh` and `mole_http` with a local
network (192.168.33.0/24) that connects them.

`mole_ssh` runs a ssh server listening on port `22`.
This port is published on the local computer using port `22122`, so ssh
connections can be made using address `127.0.0.1:22122`.
All ssh connection to `mole_ssh` should be done using the user `mole` and the
key file located on `test-env/key`

`mole_http` runs a http server listening on port `80`.
The http server responds only to requests to `http://192.168.33.11/`.

### Teardown

```sh
$ make rm-test-env
```

This will destroy both of the containers that was built by running
`make test-env`: `mole_ssh` and `mole_http`.

The ssh authentication key files, `test-env/key` and `test-env/key,pub` will
**not** be deleted.

## How to use the test environment and mole together

```sh
$ make test-env
<lots of output messages here>
$ mole -insecure -local :21112 -remote 192.168.33.11:80 -server mole@127.0.0.1:22122 -key test-env/key
INFO[0000] listening on local address                    local_address="127.0.0.1:21112"
$ curl 127.0.0.1:21112
:)
```

NOTE: If you're wondering about the smile face, that is the response from the
http server.

## Packet Analisys

If you need to analyze the traffic going through the tunnel, the test
environment provide a handy way to sniff all traffic following the steps below:

```sh
$ make test-env
<lots of output messages here>
$ bash test-env/sniff
tcpdump: listening on eth0, link-type EN10MB (Ethernet), capture size 262144 bytes
```

Wireshark will open and should show all traffic passing through the tunnel.
