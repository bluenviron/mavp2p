
# mavp2p

[![Go Report Card](https://goreportcard.com/badge/github.com/gswly/mavp2p)](https://goreportcard.com/report/github.com/gswly/mavp2p)
[![Build Status](https://travis-ci.org/gswly/mavp2p.svg?branch=master)](https://travis-ci.org/gswly/mavp2p)

mavp2p is a flexible and efficient Mavlink proxy / bridge / router. It is used primarily for linking UAV flight controllers, connected through a serial port, with ground stations on a network, but can be used to build any kind of routing involving serial, TCP and UDP, allowing communication across different physical layers or transport layers.

This software is intended as a replacement for mavproxy in systems with limited resources (i.e. companion computers), and as a replacement for mavlink-router when flexibility is needed.

This software is based on the [**gomavlib**](https://github.com/gswly/gomavlib) library.

## Features

* Supports Mavlink 2.0 and 1.0
* Links together an arbitrary number of endpoints:
  * serial
  * UDP (client, server or broadcast mode)
  * TCP (client or server mode)
* Transports can be reached by domain name or IP

Advantages with respect to mavproxy:
* Much lower CPU and memory usage
* Arbitrary number of inputs and outputs

Advantages with respect to mavlink-router:
* Supports domain names
* Supports multiple TCP servers

## Installation

Precompiled binaries are available in the [release](https://github.com/gswly/mavp2p/releases) page.

## Usage

Receive Mavlink via serial port and transmit it via UDP:
```
mavp2p serial:/dev/ttyAMA0 udpc:1.2.3.4:5600
```

Receive Mavlink via UDP broadcast and transmit it via TCP:
```
mavp2p udpb:192.168.7.255 tcpc:1.2.3.4:5600
```

Full command-line usage:
```
usage: mavp2p [<flags>] <endpoints>...

mavp2p v0.0.0 (fffffff)

Link together specified Mavlink endpoints.

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Args:
  <endpoints>  space-separated list of endpoints. possible endpoints are:

               serial:port:baudrate (serial)

               udps:listen_ip:port (udp, server mode)

               udpc:dest_ip:port (udp, client mode)

               udpb:broadcast_ip:port (udp, broadcast mode)

               tcps:listen_ip:port (tcp, server mode)

               tcpc:dest_ip:port (tcp, client mode)
```

## Links

Similar software
* https://github.com/ArduPilot/MAVProxy
* https://github.com/intel/mavlink-router

Mavlink references
* https://mavlink.io/en/
