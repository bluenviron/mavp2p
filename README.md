
# mavp2p

[![Go Report Card](https://goreportcard.com/badge/github.com/gswly/mavp2p)](https://goreportcard.com/report/github.com/gswly/mavp2p)
[![Build Status](https://travis-ci.org/gswly/mavp2p.svg?branch=master)](https://travis-ci.org/gswly/mavp2p)

_mavp2p_ is a flexible and efficient Mavlink proxy / bridge / router, implemented in the form of a command-line utility. It is used primarily for linking UAV flight controllers, connected through a serial port, with ground stations on a network, but can be used to build any kind of routing involving serial, TCP and UDP, allowing communication across different physical layers or transport layers.

_mavp2p_ can replace _mavproxy_ in systems with limited resources (for instance companion computers), and _mavlink-router_ when more flexibility is needed.

This project makes use of the [**gomavlib**](https://github.com/gswly/gomavlib) library, a full-featured Mavlink library.

## Features

Main features:
* Links together an arbitrary number of different types of endpoints:
  * serial
  * UDP (client, server or broadcast mode)
  * TCP (client or server mode)
* Supports Mavlink 2.0 and 1.0, supports any dialect
* Emits heartbeats
* Requests streams to Ardupilot devices and blocks stream requests from ground stations
* Supports domain names in place of IPs
* Reconnects to TCP/UDP servers when disconnected, removes inactive TCP/UDP clients
* Multiplatform, available for multiple operating systems (Linux, Windows) and architectures (arm6, arm7, armhf, amd64), does not depend on libc and therefore is compatible with lightweight distros (Alpine Linux)

Advantages with respect to _mavproxy_:
* Does not require python nor any interpreter
* Much lower CPU and memory usage
* Supports an arbitrary number of inputs and outputs
* Logs can be disabled, resulting in no disk I/O
* UDP clients are removed when inactive

Advantages with respect to _mavlink-router_:
* Supports domain names
* Supports multiple TCP servers
* UDP clients are removed when inactive
* Supports automatic stream requests to Ardupilot devices

## Installation

Precompiled binaries are available in the [release](https://github.com/gswly/mavp2p/releases) page.

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

Create a server that links together all UDP endpoints that connects to it:
```
./mavp2p udps:0.0.0.0:5600
```

Full command-line usage:
```
usage: mavp2p [<flags>] [<endpoints>...]

mavp2p v0.0.0

Link together Mavlink endpoints.

Flags:
      --help                     Show context-sensitive help (also try
                                 --help-long and --help-man).
      --version                  print mavp2p version
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
                 endpoint is required. Possible endpoints types are:

                 udps:listen_ip:port (udp, server mode)

                 udpc:dest_ip:port (udp, client mode)

                 udpb:broadcast_ip:port (udp, broadcast mode)

                 tcps:listen_ip:port (tcp, server mode)

                 tcpc:dest_ip:port (tcp, client mode)

                 serial:port:baudrate (serial)

```

## Links

Similar software
* https://github.com/ArduPilot/MAVProxy
* https://github.com/intel/mavlink-router

Mavlink references
* https://mavlink.io/en/
