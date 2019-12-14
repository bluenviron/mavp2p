package main

import (
	"fmt"
	"github.com/aler9/gomavlib"
	"github.com/aler9/gomavlib/dialects/common"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"
)

var Version string = "v0.0.0"

var reArgs = regexp.MustCompile("^([a-z]+):(.+)$")

type endpointType struct {
	args string
	desc string
	make func(args string) gomavlib.EndpointConf
}

var endpointTypes = map[string]endpointType{
	"serial": {
		"port:baudrate",
		"serial",
		func(args string) gomavlib.EndpointConf {
			return gomavlib.EndpointSerial{args}
		},
	},
	"udps": {
		"listen_ip:port",
		"udp, server mode",
		func(args string) gomavlib.EndpointConf {
			return gomavlib.EndpointUdpServer{args}
		},
	},
	"udpc": {
		"dest_ip:port",
		"udp, client mode",
		func(args string) gomavlib.EndpointConf {
			return gomavlib.EndpointUdpClient{args}
		},
	},
	"udpb": {
		"broadcast_ip:port",
		"udp, broadcast mode",
		func(args string) gomavlib.EndpointConf {
			return gomavlib.EndpointUdpBroadcast{BroadcastAddress: args}
		},
	},
	"tcps": {
		"listen_ip:port",
		"tcp, server mode",
		func(args string) gomavlib.EndpointConf {
			return gomavlib.EndpointTcpServer{args}
		},
	},
	"tcpc": {
		"dest_ip:port",
		"tcp, client mode",
		func(args string) gomavlib.EndpointConf {
			return gomavlib.EndpointTcpClient{args}
		},
	},
}

func initError(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+msg+"\n", args...)
	os.Exit(1)
}

func main() {
	kingpin.CommandLine.Help = "mavp2p " + Version + "\n\n" +
		"Link together Mavlink endpoints."

	version := kingpin.Flag("version", "print mavp2p version").Bool()

	quiet := kingpin.Flag("quiet", "suppress info messages").Short('q').Bool()
	print := kingpin.Flag("print", "print routed frames").Bool()
	printErrorsSingularly := kingpin.Flag("print-errors", "print parse errors singularly, instead of printing only their quantity every 5 seconds").Bool()

	hbDisable := kingpin.Flag("hb-disable", "disable heartbeats").Bool()
	hbVersion := kingpin.Flag("hb-version", "set mavlink version of heartbeats").Default("1").Enum("1", "2")
	hbSystemId := kingpin.Flag("hb-systemid", "set system id of heartbeats"+
		"it is recommended to set a different system id for each router in the network.").Default("125").Int()
	hbPeriod := kingpin.Flag("hb-period", "set period of heartbeats").Default("5").Int()

	streamReqDisable := kingpin.Flag("streamreq-disable", "do not request streams to Ardupilot devices, "+
		"that need an explicit request in order to emit telemetry streams. "+
		"this task is usually delegated to the router, in order to avoid conflicts when "+
		"multiple ground stations are active.").Bool()
	streamReqFrequency := kingpin.Flag("streamreq-frequency", "set the stream frequency to request").Default("4").Int()

	desc := "Space-separated list of endpoints. At least one endpoint is required. " +
		"Possible endpoints types are:\n\n"
	for k, etype := range endpointTypes {
		desc += fmt.Sprintf("%s:%s (%s)\n\n", k, etype.args, etype.desc)
	}
	endpoints := kingpin.Arg("endpoints", desc).Strings()

	kingpin.Parse()

	// print version
	if *version == true {
		fmt.Println("mavp2p " + Version)
		os.Exit(0)
	}

	// print usage if no args are provided
	if len(os.Args) <= 1 {
		kingpin.Usage()
		os.Exit(1)
	}

	if len(*endpoints) < 1 {
		initError("at least one endpoint is required")
	}

	var econfs []gomavlib.EndpointConf
	for _, e := range *endpoints {
		matches := reArgs.FindStringSubmatch(e)
		if matches == nil {
			initError("invalid endpoint: %s", e)
		}
		key, args := matches[1], matches[2]

		etype, ok := endpointTypes[key]
		if ok == false {
			initError("invalid endpoint: %s", e)
		}

		econfs = append(econfs, etype.make(args))
	}

	// decode/encode only a minimal set of messages.
	// other messages change too frequently and cannot be integrated into a static tool.
	msgs := []gomavlib.Message{}
	if *hbDisable == false || *streamReqDisable == false {
		msgs = append(msgs, &common.MessageHeartbeat{})
	}
	if *streamReqDisable == false {
		msgs = append(msgs, &common.MessageRequestDataStream{})
	}
	dialect, err := gomavlib.NewDialect(3, msgs)
	if err != nil {
		initError(err.Error())
	}

	node, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints: econfs,
		Dialect:   dialect,
		OutVersion: func() gomavlib.Version {
			if *hbVersion == "2" {
				return gomavlib.V2
			}
			return gomavlib.V1
		}(),
		OutSystemId:            byte(*hbSystemId),
		HeartbeatDisable:       *hbDisable,
		HeartbeatPeriod:        (time.Duration(*hbPeriod) * time.Second),
		StreamRequestEnable:    !*streamReqDisable,
		StreamRequestFrequency: *streamReqFrequency,
	})
	if err != nil {
		initError(err.Error())
	}
	defer node.Close()

	eh, err := newErrorHandler(*printErrorsSingularly)
	if err != nil {
		initError(err.Error())
	}

	nh, err := newNodeHandler()
	if err != nil {
		initError(err.Error())
	}

	if *quiet == true {
		log.SetOutput(ioutil.Discard)
	}

	log.Printf("mavp2p %s", Version)
	log.Printf("router started with %d endpoints", len(econfs))

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
				evt.SystemId, evt.ComponentId)

		case *gomavlib.EventFrame:
			if *print == true {
				fmt.Printf("%#v, %#v\n", evt.Frame, evt.Message())
			}

			nh.onEventFrame(evt)

			// if automatic stream requests are enabled, block manual stream requests
			if *streamReqDisable == false {
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
}
