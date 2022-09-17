// main executable.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/aler9/gomavlib"
	"github.com/aler9/gomavlib/pkg/dialect"
	"github.com/aler9/gomavlib/pkg/dialects/common"
	"github.com/aler9/gomavlib/pkg/message"
	"gopkg.in/alecthomas/kingpin.v2"
)

var version = "v0.0.0"

var (
	reArgs   = regexp.MustCompile("^([a-z]+):(.+)$")
	reSerial = regexp.MustCompile("^(.+?):([0-9]+)$")
)

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
			matches := reSerial.FindStringSubmatch(args)
			if matches == nil {
				return nil, fmt.Errorf("invalid address")
			}

			dev := matches[1]
			baud, _ := strconv.Atoi(matches[2])

			return gomavlib.EndpointSerial{
				Device: dev,
				Baud:   baud,
			}, nil
		},
	},
	"udps": {
		"listen_ip:port",
		"udp, server mode",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointUDPServer{Address: args}, nil
		},
	},
	"udpc": {
		"dest_ip:port",
		"udp, client mode",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointUDPClient{Address: args}, nil
		},
	},
	"udpb": {
		"broadcast_ip:port",
		"udp, broadcast mode",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointUDPBroadcast{BroadcastAddress: args}, nil
		},
	},
	"tcps": {
		"listen_ip:port",
		"tcp, server mode",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointTCPServer{Address: args}, nil
		},
	},
	"tcpc": {
		"dest_ip:port",
		"tcp, client mode",
		func(args string) (gomavlib.EndpointConf, error) {
			return gomavlib.EndpointTCPClient{Address: args}, nil
		},
	},
}

func run() error {
	kingpin.CommandLine.Help = "mavp2p " + version + "\n\n" +
		"Link together Mavlink endpoints."

	argVersion := kingpin.Flag("version", "print version").Bool()

	argQuiet := kingpin.Flag("quiet", "suppress info messages").Short('q').Bool()
	argPrint := kingpin.Flag("print", "print routed frames").Bool()
	argPrintSingleErrors := kingpin.Flag("print-errors",
		"print parse errors singularly, instead of printing only their quantity every 5 seconds").Bool()

	argHbDisable := kingpin.Flag("hb-disable", "disable heartbeats").Bool()
	argHbVersion := kingpin.Flag("hb-version", "set mavlink version of heartbeats").Default("1").Enum("1", "2")
	argHbSystemID := kingpin.Flag("hb-systemid", "set system id of heartbeats"+
		"it is recommended to set a different system id for each router in the network.").Default("125").Int()
	argHbPeriod := kingpin.Flag("hb-period", "set period of heartbeats").Default("5").Int()

	argStreamReqDisable := kingpin.Flag("streamreq-disable", "do not request streams to Ardupilot devices, "+
		"that need an explicit request in order to emit telemetry streams. "+
		"this task is usually delegated to the router, in order to avoid conflicts when "+
		"multiple ground stations are active.").Bool()
	argStreamReqFrequency := kingpin.Flag("streamreq-frequency", "set the stream frequency to request").Default("4").Int()

	desc := "Space-separated list of endpoints. At least one endpoint is required. " +
		"Possible endpoints types are:\n\n"
	for k, etype := range endpointTypes {
		desc += fmt.Sprintf("%s:%s (%s)\n\n", k, etype.args, etype.desc)
	}
	endpoints := kingpin.Arg("endpoints", desc).Strings()

	kingpin.Parse()

	// print version
	if *argVersion {
		fmt.Println(version)
		return nil
	}

	// print usage if no args are provided
	if len(os.Args) <= 1 {
		kingpin.Usage()
		os.Exit(1)
	}

	if len(*endpoints) < 1 {
		return fmt.Errorf("at least one endpoint is required")
	}

	econfs := make([]gomavlib.EndpointConf, len(*endpoints))
	for i, e := range *endpoints {
		matches := reArgs.FindStringSubmatch(e)
		if matches == nil {
			return fmt.Errorf("invalid endpoint: %s", e)
		}
		key, args := matches[1], matches[2]

		etype, ok := endpointTypes[key]
		if !ok {
			return fmt.Errorf("invalid endpoint: %s", e)
		}

		conf, err := etype.make(args)
		if err != nil {
			return err
		}
		econfs[i] = conf
	}

	// decode/encode only a minimal set of messages.
	// other messages change too frequently and cannot be integrated into a static tool.
	msgs := []message.Message{}
	if !*argHbDisable || !*argStreamReqDisable {
		msgs = append(msgs, &common.MessageHeartbeat{})
	}
	if !*argStreamReqDisable {
		msgs = append(msgs, &common.MessageRequestDataStream{})
	}
	dialect := &dialect.Dialect{3, msgs} //nolint:govet

	node, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints: econfs,
		Dialect:   dialect,
		OutVersion: func() gomavlib.Version {
			if *argHbVersion == "2" {
				return gomavlib.V2
			}
			return gomavlib.V1
		}(),
		OutSystemID:            byte(*argHbSystemID),
		HeartbeatDisable:       *argHbDisable,
		HeartbeatPeriod:        (time.Duration(*argHbPeriod) * time.Second),
		StreamRequestEnable:    !*argStreamReqDisable,
		StreamRequestFrequency: *argStreamReqFrequency,
	})
	if err != nil {
		return err
	}
	defer node.Close()

	eh, err := newErrorHandler(*argPrintSingleErrors)
	if err != nil {
		return err
	}

	nh, err := newNodeHandler()
	if err != nil {
		return err
	}

	if *argQuiet {
		log.SetOutput(io.Discard)
	}

	log.Printf("mavp2p %s", version)
	log.Printf("router started with %d %s",
		len(econfs),
		func() string {
			if len(econfs) == 1 {
				return "endpoint"
			}
			return "endpoints"
		}())

	go eh.run()
	go nh.run()

	for e := range node.Events() {
		switch evt := e.(type) {
		case *gomavlib.EventChannelOpen:
			log.Printf("channel opened: %s", evt.Channel)

		case *gomavlib.EventChannelClose:
			log.Printf("channel closed: %s", evt.Channel)
			nh.onEventChannelClose(evt)

		case *gomavlib.EventStreamRequested:
			log.Printf("stream requested to chan=%s sid=%d cid=%d", evt.Channel,
				evt.SystemID, evt.ComponentID)

		case *gomavlib.EventFrame:
			if *argPrint {
				fmt.Printf("%#v, %#v\n", evt.Frame, evt.Message())
			}

			nh.onEventFrame(evt)

			// if automatic stream requests are enabled, block manual stream requests
			if !*argStreamReqDisable {
				if _, ok := evt.Message().(*common.MessageRequestDataStream); ok {
					continue
				}
			}

			// route message to every other channel
			node.WriteFrameExcept(evt.Channel, evt.Frame)

		case *gomavlib.EventParseError:
			eh.onEventError(evt)
		}
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
		os.Exit(1)
	}
}
