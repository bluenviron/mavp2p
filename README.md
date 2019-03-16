
# mavp2p

[![Go Report Card](https://goreportcard.com/badge/github.com/gswly/mavp2p)](https://goreportcard.com/report/github.com/gswly/mavp2p)
[![Build Status](https://travis-ci.org/gswly/mavp2p.svg?branch=master)](https://travis-ci.org/gswly/mavp2p)

mavp2p is a Mavlink proxy / bridge / router focusing on efficiency and flexibility. It is used primarily for linking UAV flight controllers, connected through a serial port, with ground stations on a network, but can be used to build any kind of routing involving serial, TCP and UDP, linking together multiple devices across different physical layers or transport layers.

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

TODO

## Usage

TODO

Usage examples:

TODO

## Links

Similar software
* https://github.com/ArduPilot/MAVProxy
* https://github.com/intel/mavlink-router

Mavlink references
* https://mavlink.io/en/
