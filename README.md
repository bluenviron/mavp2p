
# mavp2p

[![Test](https://github.com/bluenviron/mavp2p/workflows/test/badge.svg)](https://github.com/bluenviron/mavp2p/actions?query=workflow:test)
[![Lint](https://github.com/bluenviron/mavp2p/workflows/lint/badge.svg)](https://github.com/bluenviron/mavp2p/actions?query=workflow:lint)
[![Release](https://img.shields.io/github/v/release/bluenviron/mavp2p)](https://github.com/bluenviron/mavp2p/releases)
[![Docker Hub](https://img.shields.io/badge/docker-aler9/mavp2p-blue)](https://hub.docker.com/r/aler9/mavp2p)

_mavp2p_ is a flexible and efficient Mavlink proxy / bridge / router, implemented in the form of a command-line utility. It is used primarily to link UAV flight controllers, connected through a serial port, with ground stations on a network, but can be used to build any kind of routing involving serial, TCP and UDP, allowing communication across different physical layers or transport layers.

This project makes use of the [**gomavlib**](https://github.com/bluenviron/gomavlib) library, a full-featured Mavlink library.

Features:

* Link together an arbitrary number of different kinds of endpoints:
  * serial
  * UDP (client, server or broadcast mode)
  * TCP (client or server mode)
* Support Mavlink 2.0 and 1.0, support any dialect
* Emit heartbeats
* Request streams to Ardupilot devices and block stream requests from ground stations
* Support domain names in place of IPs
* Reconnect to TCP/UDP servers when disconnected, remove inactive TCP/UDP clients
* Multiplatform, available for multiple operating systems (Linux, Windows) and architectures (arm6, arm7, arm64, amd64), does not depend on libc and therefore is compatible with lightweight distros (Alpine Linux)

## Important announcement

my main open source projects are being transferred to the [bluenviron organization](https://github.com/bluenviron), in order to allow the community to maintain and evolve the code regardless of my personal availability.

In the next months, the repository name will be changed accordingly.

## Table of contents

* [Installation](#installation)
* [Usage](#usage)
* [Comparison with similar software](#comparison-with-similar-software)
* [Full command-line usage](#full-command-line-usage)
* [Links](#links)

## Installation

Download and extract a precompiled binary from the [release page](https://github.com/bluenviron/mavp2p/releases).

If you want to use Docker, there's a image available at `aler9/mavp2p`:

```
docker run --rm -it --network=host -e COLUMNS=$COLUMNS aler9/mavp2p
```

## Usage

Link a serial port with a UDP endpoint in client mode:

```
./mavp2p serial:/dev/ttyAMA0:57600 udpc:1.2.3.4:5600
```

Link a serial port with a UDP endpoint in server mode:

```
./mavp2p serial:/dev/ttyAMA0:57600 udps:0.0.0.0:5600
```

Link a UDP endpoint in broadcast mode with a TCP endpoint in client mode:

```
./mavp2p udpb:192.168.7.255:5601 tcpc:exampleendpoint.com:5600
```

Create a server that links together all UDP endpoints that connect to it:

```
./mavp2p udps:0.0.0.0:5600
```

## Comparison with similar software

_mavp2p_ vs _mavproxy_

* Does not require python nor any interpreter
* Much lower CPU and memory usage
* Supports an arbitrary number of inputs and outputs
* Logs can be disabled, resulting in no disk I/O
* UDP clients are removed when inactive

_mavp2p_ vs _mavlink-router_

* Supports domain names
* Supports multiple TCP servers
* UDP clients are removed when inactive
* Supports automatic stream requests to Ardupilot devices

## Full command-line usage

```
usage: mavp2p [<flags>] [<endpoints>...]

mavp2p v0.0.0

Link together Mavlink endpoints.

Flags:
      --help                     Show context-sensitive help (also try
                                 --help-long and --help-man).
      --version                  print version
  -q, --quiet                    suppress info messages
      --print                    print routed frames
      --print-errors             print parse errors singularly, instead of
                                 printing only their quantity every 5 seconds
      --hb-disable               disable heartbeats
      --hb-version=1             set mavlink version of heartbeats
      --hb-systemid=125          set system id of heartbeats. it is
                                 recommended to set a different system
                                 id for each router in the network
      --hb-period=5              set period of heartbeats
      --streamreq-disable        do not request streams to Ardupilot
                                 devices, that need an explicit request
                                 in order to emit telemetry streams.
                                 this task is usually delegated to the
                                 router, in order to avoid conflicts when
                                 multiple ground stations are active
      --streamreq-frequency=4    set the stream frequency to request

Args:
  [<endpoints>]  Space-separated list of endpoints. At least one
                 endpoint is required. Possible endpoints kinds are:

                 udps:listen_ip:port (udp, server mode)

                 udpc:dest_ip:port (udp, client mode)

                 udpb:broadcast_ip:port (udp, broadcast mode)

                 tcps:listen_ip:port (tcp, server mode)

                 tcpc:dest_ip:port (tcp, client mode)

                 serial:port:baudrate (serial)

```

## Links

Related projects

* https://github.com/bluenviron/gomavlib

Similar software

* https://github.com/ArduPilot/MAVProxy
* https://github.com/intel/mavlink-router

Mavlink references

* https://mavlink.io/en/
