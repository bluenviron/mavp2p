package main

import (
	"fmt"
	"github.com/gswly/gomavlib"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"regexp"
)

var Version string = "(unknown version)"

var reArgs = regexp.MustCompile("^([a-z]+):(.+)$")

type endpointType struct {
	args string
	desc string
	make func(args string) (gomavlib.EndpointConf, error)
}

var endpointTypes = map[string]endpointType{
	"serial": {
		"port:baudrate",
		"serial",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointSerial{args}, nil
		},
	},
	"udps": {
		"listen_ip:port",
		"udp, server mode",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointUdpServer{args}, nil
		},
	},
	"udpc": {
		"dest_ip:port",
		"udp, client mode",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointUdpClient{args}, nil
		},
	},
	"udpb": {
		"broadcast_ip:port",
		"udp, broadcast mode",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointUdpBroadcast{BroadcastAddress: args}, nil
		},
	},
	"tcps": {
		"listen_ip:port",
		"tcp, server mode",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointTcpServer{args}, nil
		},
	},
	"tcpc": {
		"dest_ip:port",
		"tcp, client mode",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointTcpClient{args}, nil
		},
	},
}

func main() {
	kingpin.CommandLine.Help = "mavp2p " + Version + "\n\n" +
		"Link together specified Mavlink endpoints."

	desc := "space-separated list of endpoints. " +
		"possible endpoints are:\n\n"
	for k, etype := range endpointTypes {
		desc += fmt.Sprintf("%s:%s (%s)\n\n", k, etype.args, etype.desc)
	}
	endpoints := kingpin.Arg("endpoints", desc).Strings()

	kingpin.Parse()

	if len(*endpoints) < 2 {
		log.Fatalf("at least 2 endpoints are required.")
	}

	var econfs []gomavlib.EndpointConf
	for _, e := range *endpoints {
		matches := reArgs.FindStringSubmatch(e)
		if matches == nil {
			log.Fatalf("invalid endpoint: %s", e)
		}
		key, args := matches[1], matches[2]

		etype, ok := endpointTypes[key]
		if ok == false {
			log.Fatalf("invalid endpoint: %s", e)
		}

		e, err := etype.make(args)
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		econfs = append(econfs, e)
	}

	node, err := gomavlib.NewNode(gomavlib.NodeConf{
		Dialect:     nil,
		SystemId:    125,
		ComponentId: 1,
		Endpoints:   econfs,
	})
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	defer node.Close()

	log.Printf("router started with %d endpoints", len(econfs))

	for {
		// wait until a message is received.
		res, ok := node.Read()
		if ok == false {
			break
		}

		// print message details
		//fmt.Printf("received: id=%d, %+v\n", res.Message().GetId(), res.Message())

		// route message to every other channel
		node.WriteFrameExcept(res.Channel(), res.Frame())
	}
}
