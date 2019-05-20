package main

import (
	"fmt"
	"github.com/gswly/gomavlib"
	"github.com/gswly/gomavlib/dialects/common"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sync"
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

type remoteNode struct {
	Channel     *gomavlib.Channel
	SystemId    byte
	ComponentId byte
}

func (i remoteNode) String() string {
	return fmt.Sprintf("chan=%s sid=%d cid=%d", i.Channel, i.SystemId, i.ComponentId)
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
	printErrors := kingpin.Flag("print-errors", "print parse errors individually, instead of printing only their quantity every 5 seconds").Bool()

	hbDisable := kingpin.Flag("hb-disable", "disable heartbeats").Bool()
	hbVersion := kingpin.Flag("hb-version", "set mavlink version of heartbeats").Default("1").Enum("1", "2")
	hbSystemId := kingpin.Flag("hb-systemid", "set system id of heartbeats"+
		"it is recommended to set a different system id for each router in the network.").Default("125").Int()
	hbPeriod := kingpin.Flag("hb-period", "set period of heartbeats").Default("5").Int()

	aprsDisable := kingpin.Flag("apreqstream-disable", "do not request streams to Ardupilot devices, "+
		"that need an explicit request in order to emit telemetry streams. "+
		"this task is usually delegated to the router, to avoid conflicts when "+
		"multiple ground stations are active.").Bool()
	aprsFrequency := kingpin.Flag("apreqstream-frequency", "set the stream frequency to request").Default("4").Int()

	desc := "Space-separated list of endpoints. At least 2 endpoints are required. " +
		"Possible endpoints are:\n\n"
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

	if len(*endpoints) < 2 {
		initError("at least 2 endpoints are required")
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
	if *hbDisable == false {
		msgs = append(msgs, &common.MessageHeartbeat{})
	}
	if *aprsDisable == false {
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
		OutSystemId:      byte(*hbSystemId),
		HeartbeatDisable: *hbDisable,
		HeartbeatPeriod:  (time.Duration(*hbPeriod) * time.Second),
	})
	if err != nil {
		initError(err.Error())
	}
	defer node.Close()

	if *quiet == true {
		log.SetOutput(ioutil.Discard)
	}

	log.Printf("mavp2p %s", Version)
	log.Printf("router started with %d endpoints", len(econfs))

	errorCount := 0
	var remoteNodeMutex sync.Mutex
	remoteNodes := make(map[remoteNode]time.Time)
	requestedStreams := make(map[remoteNode]struct{})

	// print errors
	if *printErrors == false {
		go func() {
			for {
				time.Sleep(5 * time.Second)
				if errorCount > 0 {
					log.Printf("%d errors in the last 5 seconds", errorCount)
					errorCount = 0
				}
			}
		}()
	}

	// delete remote nodes after a period of inactivity
	go func() {
		for {
			time.Sleep(10 * time.Second)
			func() {
				remoteNodeMutex.Lock()
				defer remoteNodeMutex.Unlock()
				for rnode, t := range remoteNodes {
					if time.Since(t) >= (30 * time.Second) {
						delete(remoteNodes, rnode)
						log.Printf("node disappeared: %s", rnode)
					}
				}
			}()
		}
	}()

	for e := range node.Events() {
		switch evt := e.(type) {
		case *gomavlib.EventChannelOpen:
			log.Printf("channel opened: %s", evt.Channel)

		case *gomavlib.EventChannelClose:
			// delete remote nodes associated to channel
			func() {
				remoteNodeMutex.Lock()
				defer remoteNodeMutex.Unlock()
				for rnode := range remoteNodes {
					if rnode.Channel == evt.Channel {
						delete(remoteNodes, rnode)
						log.Printf("node disappeared: %s", rnode)
					}
				}
			}()
			log.Printf("channel closed: %s", evt.Channel)

		case *gomavlib.EventFrame:
			// print
			if *print == true {
				fmt.Printf("%#v, %#v\n", evt.Frame, evt.Message())
			}

			// build remoteNode
			rnode := remoteNode{
				Channel:     evt.Channel,
				SystemId:    evt.SystemId(),
				ComponentId: evt.ComponentId(),
			}

			func() {
				remoteNodeMutex.Lock()
				defer remoteNodeMutex.Unlock()

				// new remote node
				if _, ok := remoteNodes[rnode]; !ok {
					log.Printf("node appeared: %s", rnode)
				}

				// always update time
				remoteNodes[rnode] = time.Now()
			}()

			// request streams to ardupilot devices
			if *aprsDisable == false {
				// request if node is ardupilot and new
				if hb, ok := evt.Message().(*common.MessageHeartbeat); ok &&
					hb.Autopilot == common.MAV_AUTOPILOT_ARDUPILOTMEGA {
					if _, ok := requestedStreams[rnode]; !ok {
						requestedStreams[rnode] = struct{}{}

						// https://github.com/mavlink/qgroundcontrol/blob/08f400355a8f3acf1dd8ed91f7f1c757323ac182/src/FirmwarePlugin/APM/APMFirmwarePlugin.cc#L626
						streams := []common.MAV_DATA_STREAM{
							common.MAV_DATA_STREAM_RAW_SENSORS,
							common.MAV_DATA_STREAM_EXTENDED_STATUS,
							common.MAV_DATA_STREAM_RC_CHANNELS,
							common.MAV_DATA_STREAM_POSITION,
							common.MAV_DATA_STREAM_EXTRA1,
							common.MAV_DATA_STREAM_EXTRA2,
							common.MAV_DATA_STREAM_EXTRA3,
						}

						for _, stream := range streams {
							node.WriteMessageTo(evt.Channel, &common.MessageRequestDataStream{
								TargetSystem:    evt.SystemId(),
								TargetComponent: evt.ComponentId(),
								ReqStreamId:     uint8(stream),
								ReqMessageRate:  uint16(*aprsFrequency),
								StartStop:       1,
							})
						}
					}
				}

				// stop requests from ground stations
				if _, ok := evt.Message().(*common.MessageRequestDataStream); ok {
					continue
				}
			}

			// route message to every other channel
			node.WriteFrameExcept(evt.Channel, evt.Frame)

		case *gomavlib.EventParseError:
			if *printErrors == true {
				log.Printf("ERR: %s", evt.Error)
			} else {
				errorCount++
			}
		}
	}
}
