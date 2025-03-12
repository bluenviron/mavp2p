
# mavp2p

[![Test](https://github.com/bluenviron/mavp2p/actions/workflows/test.yml/badge.svg)](https://github.com/bluenviron/mavp2p/actions/workflows/test.yml)
[![Lint](https://github.com/bluenviron/mavp2p/actions/workflows/lint.yml/badge.svg)](https://github.com/bluenviron/mavp2p/actions/workflows/lint.yml)
[![CodeCov](https://codecov.io/gh/bluenviron/mavp2p/branch/main/graph/badge.svg)](https://app.codecov.io/gh/bluenviron/mavp2p/tree/main)
[![Release](https://img.shields.io/github/v/release/bluenviron/mavp2p)](https://github.com/bluenviron/mavp2p/releases)
[![Docker Hub](https://img.shields.io/badge/docker-bluenviron/mavp2p-blue)](https://hub.docker.com/r/bluenviron/mavp2p)

_mavp2p_ is a flexible and efficient Mavlink proxy / bridge / router, implemented in the form of a command-line utility. It is used primarily to link UAV flight controllers, connected through a serial port, with ground stations on a network, but can be used to build any kind of routing involving serial, TCP and UDP, allowing communication across different physical layers or transport layers.

This project makes use of the [**gomavlib**](https://github.com/bluenviron/gomavlib) library, a full-featured Mavlink library.

Features:

* Link together an arbitrary number of different kinds of endpoints:
  * Serial
  * UDP (client, server or broadcast mode)
  * TCP (client or server mode)
* Support Mavlink 2.0 and 1.0, support any dialect
* Emit heartbeats
* Automatically request streams to Ardupilot devices and block stream requests from ground stations
* Route messages by target system ID / component ID
* Use domain names in place of IPs
* Reconnect to TCP/UDP servers when disconnected, remove inactive TCP/UDP clients
* Dump telemetry to disk
* Multiplatform, available for multiple operating systems (Linux, Windows) and architectures (arm6, arm7, arm64, amd64), independent from libc and compatible with lightweight distros (Alpine Linux)

## Table of contents

* [Installation](#installation)
  * [Standalone binary](#standalone-binary)
  * [Docker image](#docker-image)
  * [OpenWrt binary](#openwrt-binary)
* [Usage](#usage)
* [Connecting popular software](#connecting-popular-software)
  * [QGroundControl](#qgroundcontrol)
  * [Mission Planner](#mission-planner)
  * [PX4 Simulator](#px4-simulator)
* [Comparison with similar software](#comparison-with-similar-software)
* [Full command-line usage](#full-command-line-usage)
* [Compile from source](#compile-from-source)
  * [Standard](#standard)
  * [OpenWrt](#openwrt)
  * [Cross compile](#cross-compile)
* [Specifications](#specifications)
* [Links](#links)

## Installation

There are several installation methods available: standalone binary, Docker image and OpenWRT binary.

### Standalone binary

Download and extract a standalone binary from the [release page](https://github.com/bluenviron/mavp2p/releases) that corresponds to your operating system and architecture.

### Docker image

There's a image available at `bluenviron/mavp2p`:

```
docker run --rm -it --network=host bluenviron/mavp2p
```

### OpenWrt binary

If the architecture of the OpenWrt device is amd64, armv6, armv7 or arm64, use the [standalone binary method](#standalone-binary) and download a Linux binary that corresponds to your architecture.

Otherwise, [compile the software from source](#openwrt).

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

Dump telemetry to disk:

```
./mavp2p udps:0.0.0.0:5600 --dump --dump-path="dump/2006-01-02_15-04-05.tlog"
```

## Connecting popular software

### QGroundControl

1. Make sure that mavp2p is started with a UDP server endpoint:

   ```
   ./mavp2p udps:0.0.0.0:5600
   ```

2. Install and open QGroundControl.

3. Open _Application Settings_, click on _Comm Links_, _Add_. Fill the form with:
   * Name: some name
   * Type: UDP
   * Port: 0
   * Server Addresses: mavp2p-ip:5600 (then click _Add Server_)

4. Click _Ok_. Click _Connect_.

### Mission Planner

1. Make sure that mavp2p is started with a TCP server endpoint:

   ```
   ./mavp2p tcps:0.0.0.0:5601
   ```

2. Install and open Mission Planner.

3. In the top right corner of the Window, from the select menu, change the value from "AUTO" to "TCP". Then click _Connect_. Insert the mavp2p IP and port.

### PX4 Simulator

1. Make sure that mavp2p is started with a UDP server endpoint on port 14550:

   ```
   ./mavp2p udps:0.0.0.0:14550
   ```

2. [Install PX4 Simulator](https://docs.px4.io/main/en/simulation/).

3. If mavp2p and PX4 are on two different machines, enable Mavlink broadcasting by opening `PX4-Autopilot/ROMFS/px4fmu_common/init.d-posix/px4-rc.mavlink`, finding `mavlink start` and adding the `-p` flag:

   ```
   mavlink start -x -u $udp_gcs_port_local -r 4000000 -f -p
   ```

4. Start PX4 Simulator:

   ```
   make px4_sitl gz_x500
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
Usage: mavp2p [<endpoints> ...] [flags]

mavp2p v0.0.0

Arguments:
  [<endpoints> ...]    Space-separated list of endpoints. At least one endpoint is required. Possible endpoints types are:

                       serial:port:baudrate (serial)

                       udps:listen_ip:port (udp, server mode)

                       udpc:dest_ip:port (udp, client mode)

                       udpb:broadcast_ip:port (udp, broadcast mode)

                       tcps:listen_ip:port (tcp, server mode)

                       tcpc:dest_ip:port (tcp, client mode)

Flags:
  -h, --help                                         Show context-sensitive help.
      --version                                      Print version.
  -q, --quiet                                        Suppress info messages.
      --print                                        Print routed frames.
      --print-errors                                 Print parse errors singularly, instead of printing only their quantity every 5 seconds.
      --read-timeout=10s                             Timeout of read operations.
      --write-timeout=10s                            Timeout of write operations.
      --idle-timeout=60s                             Disconnect idle connections after a timeout.
      --hb-disable                                   Disable heartbeats.
      --hb-version=1                                 Mavlink version of heartbeats.
      --hb-systemid=125                              System ID of heartbeats. It is recommended to set a different system id for each router in the network.
      --hb-componentid=191                           Component ID of heartbeats.
      --hb-period=5                                  Period of heartbeats.
      --streamreq-disable                            Do not request streams to Ardupilot devices, that need an explicit request in order to emit telemetry streams. This task
                                                     is usually delegated to the router, in order to avoid conflicts when multiple ground stations are active.
      --streamreq-frequency=4                        Stream frequency to request.
      --dump                                         Dump telemetry to disk
      --dump-path="dump/2006-01-02_15-04-05.tlog"    Path of dump segments, in Golang's time.Format() format
      --dump-duration=1h                             Maximum duration of each dump segment
```

## Compile from source

### Standard

Install git and Go &ge; 1.21. Clone the repository, enter into the folder and start the building process:

```sh
git clone https://github.com/bluenviron/mavp2p
cd mavp2p
CGO_ENABLED=0 go build .
```

The command will produce the `mavp2p` binary.

### OpenWrt

The compilation procedure is the same as the standard one. On the OpenWrt device, install git and Go:

```sh
opkg update
opkg install golang git git-http
```

Clone the repository, enter into the folder and start the building process:

```sh
git clone https://github.com/bluenviron/mavp2p
cd mavp2p
CGO_ENABLED=0 go build .
```

The command will produce the `mavp2p` binary.

If the OpenWrt device doesn't have enough resources to compile, you can [cross compile](#cross-compile) from another machine.

### Cross compile

Cross compilation allows to build an executable for a target machine from another machine with different operating system or architecture. This is useful in case the target machine doesn't have enough resources for compilation or if you don't want to install the compilation dependencies on it.

On the machine you want to use to compile, install git and Go &ge; 1.21. Clone the repository, enter into the folder and start the building process:

```sh
git clone https://github.com/bluenviron/mavp2p
cd mavp2p
CGO_ENABLED=0 GOOS=my_os GOARCH=my_arch go build .
```

Replace `my_os` and `my_arch` with the operating system and architecture of your target machine. A list of all supported combinations can be obtained with:

```sh
go tool dist list
```

For instance:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build .
```

In case of the `arm` architecture, there's an additional flag available, `GOARM`, that allows to set the ARM version:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GOARM=7 go build .
```

In case of the `mips` architecture, there's an additional flag available, `GOMIPS`, that allows to set additional parameters:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat go build .
```

The command will produce the `mavp2p` binary.

## Specifications

* [Mavlink specifications](https://github.com/bluenviron/gomavlib#specifications)

## Links

Related projects

* [gomavlib](https://github.com/bluenviron/gomavlib)

Similar software

* [MavProxy](https://github.com/ArduPilot/MAVProxy)
* [mavlink-router](https://github.com/intel/mavlink-router)
