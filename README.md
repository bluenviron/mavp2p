
# mavp2p

[![Test](https://github.com/bluenviron/mavp2p/workflows/test/badge.svg)](https://github.com/bluenviron/mavp2p/actions?query=workflow:test)
[![Lint](https://github.com/bluenviron/mavp2p/workflows/lint/badge.svg)](https://github.com/bluenviron/mavp2p/actions?query=workflow:lint)
[![CodeCov](https://codecov.io/gh/bluenviron/mavp2p/branch/main/graph/badge.svg)](https://app.codecov.io/gh/bluenviron/mavp2p/branch/main)
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
* Automatically request streams to Ardupilot devices and block stream requests from ground stations
* Route messages by target system ID / component ID
* Support domain names in place of IPs
* Reconnect to TCP/UDP servers when disconnected, remove inactive TCP/UDP clients
* Multiplatform, available for multiple operating systems (Linux, Windows) and architectures (arm6, arm7, arm64, amd64), does not depend on libc and therefore is compatible with lightweight distros (Alpine Linux)

## Table of contents

* [Installation](#installation)
* [Usage](#usage)
* [Comparison with similar software](#comparison-with-similar-software)
* [Full command-line usage](#full-command-line-usage)
* [Standards](#standards)
* [Links](#links)

## Installation

There are several installation methods available: standalone binary, Docker image and OpenWRT package.

### Standalone binary

Download and extract a standalone binary from the [release page](https://github.com/bluenviron/mavp2p/releases).

### Docker image

There's a image available at `aler9/mavp2p`:

```
docker run --rm -it --network=host -e COLUMNS=$COLUMNS aler9/mavp2p
```

### OpenWRT package

1. In a x86 Linux system, download the OpenWRT SDK corresponding to the wanted OpenWRT version and target from the [OpenWRT website](https://downloads.openwrt.org/releases/) and extract it.

2. Open a terminal in the SDK folder and setup the SDK:

   ```
   ./scripts/feeds update -a
   ./scripts/feeds install -a
   make defconfig
   ```

3. Download the server Makefile and set the server version inside the file:

   ```
   mkdir package/mavp2p
   wget -O package/mavp2p/Makefile https://raw.githubusercontent.com/bluenviron/mavp2p/main/openwrt.mk
   sed -i "s/v0.0.0/$(git ls-remote --tags --sort=v:refname https://github.com/bluenviron/mavp2p | tail -n1 | sed 's/.*\///; s/\^{}//')/" package/mavp2p/Makefile
   ```

4. Compile the server:

   ```
   make package/mavp2p/compile -j$(nproc)
   ```

5. Transfer the .ipk file from `bin/packages/*/base` to the OpenWRT system and install it with:

   ```
   opkg install [ipk-file-name].ipk
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

## Standards

* [Mavlink standards](https://github.com/bluenviron/gomavlib#standards)

## Links

Related projects

* [gomavlib](https://github.com/bluenviron/gomavlib)

Similar software

* [MavProxy](https://github.com/ArduPilot/MAVProxy)
* [mavlink-router](https://github.com/intel/mavlink-router)
