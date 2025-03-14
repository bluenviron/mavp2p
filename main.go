// main executable.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialect"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"github.com/bluenviron/gomavlib/v3/pkg/message"
	"github.com/bluenviron/mavp2p/pkg/dumper"
	"github.com/bluenviron/mavp2p/pkg/errorman"
	"github.com/bluenviron/mavp2p/pkg/messageman"
)

var version = "v0.0.0"

var (
	reArgs   = regexp.MustCompile("^([a-z]+):(.+)$")
	reSerial = regexp.MustCompile("^(.+?):([0-9]+)$")
)

// decode/encode only a minimal set of messages.
// other messages change too frequently and cannot be integrated into a static tool.
func generateDialect(hbDisable bool, streamreqDisable bool) *dialect.Dialect {
	msgs := []message.Message{}

	// add all messages with the TargetSystem and TargetComponent fields
	var zero reflect.Value
	for _, msg := range common.Dialect.Messages {
		rv := reflect.ValueOf(msg).Elem()
		if rv.FieldByName("TargetSystem") != zero && rv.FieldByName("TargetComponent") != zero {
			msgs = append(msgs, msg)
		}
	}

	if !hbDisable || !streamreqDisable {
		msgs = append(msgs, &common.MessageHeartbeat{})
	}

	return &dialect.Dialect{Version: 3, Messages: msgs}
}

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

func generateEndpointConfs(endpoints []string) ([]gomavlib.EndpointConf, error) {
	if len(endpoints) < 1 {
		return nil, fmt.Errorf("at least one endpoint is required")
	}

	econfs := make([]gomavlib.EndpointConf, len(endpoints))

	for i, e := range endpoints {
		matches := reArgs.FindStringSubmatch(e)
		if matches == nil {
			return nil, fmt.Errorf("invalid endpoint: %s", e)
		}
		key, args := matches[1], matches[2]

		etype, ok := endpointTypes[key]
		if !ok {
			return nil, fmt.Errorf("invalid endpoint: %s", e)
		}

		conf, err := etype.make(args)
		if err != nil {
			return nil, err
		}
		econfs[i] = conf
	}

	return econfs, nil
}

var cli struct {
	Version            bool `help:"Print version."`
	Quiet              bool `short:"q" help:"Suppress info messages."`
	Print              bool `help:"Print routed frames."`
	PrintErrors        bool
	ReadTimeout        time.Duration `help:"Timeout of read operations." default:"10s"`
	WriteTimeout       time.Duration `help:"Timeout of write operations." default:"10s"`
	IdleTimeout        time.Duration `help:"Disconnect idle connections after a timeout." default:"60s"`
	HbDisable          bool          `help:"Disable heartbeats."`
	HbVersion          int           `enum:"1,2" help:"Mavlink version of heartbeats." default:"1"`
	HbSystemid         int           `default:"125"`
	HbComponentid      int           `help:"Component ID of heartbeats." default:"191"`
	HbPeriod           int           `help:"Period of heartbeats." default:"5"`
	StreamreqDisable   bool
	StreamreqFrequency int           `help:"Stream frequency to request." default:"4"`
	Dump               bool          `help:"Dump telemetry to disk"`
	DumpPath           string        `default:"dump/2006-01-02_15-04-05.tlog"`
	DumpDuration       time.Duration `help:"Maximum duration of each dump segment" default:"1h"`
	Endpoints          []string      `arg:"" optional:""`
}

type program struct {
	ctx        context.Context
	ctxCancel  func()
	wg         sync.WaitGroup
	node       *gomavlib.Node
	errorMan   *errorman.Manager
	messageMan *messageman.Manager
	dumper     *dumper.Dumper
}

func newProgram(args []string) (*program, error) {
	parser, err := kong.New(&cli,
		kong.Description("mavp2p "+version),
		kong.UsageOnError(),
		kong.ValueFormatter(func(value *kong.Value) string {
			switch value.Name {
			case "print-errors":
				return "Print parse errors singularly, instead of printing only their quantity every 5 seconds."

			case "hb-systemid":
				return "System ID of heartbeats. It is recommended to set a different system id for each router in the network."

			case "streamreq-disable":
				return "Do not request streams to Ardupilot devices," +
					" that need an explicit request in order to emit telemetry streams." +
					" This task is usually delegated to the router," +
					" in order to avoid conflicts when multiple ground stations are active."

			case "endpoints":
				desc := "Space-separated list of endpoints. At least one endpoint is required. " +
					"Possible endpoints types are:\n\n"
				for k, etype := range endpointTypes {
					desc += fmt.Sprintf("%s:%s (%s)\n\n", k, etype.args, etype.desc)
				}
				return desc

			case "dump-path":
				return "Path of dump segments, in Golang's time.Format() format"

			default:
				return kong.DefaultHelpValueFormatter(value)
			}
		}))
	if err != nil {
		return nil, err
	}

	kongCtx, err := parser.Parse(args)
	if err != nil {
		return nil, err
	}

	if cli.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	// print usage if no args are provided
	if len(os.Args) <= 1 {
		kongCtx.PrintUsage(false) //nolint:errcheck
		os.Exit(1)
	}

	endpointConfs, err := generateEndpointConfs(cli.Endpoints)
	if err != nil {
		return nil, err
	}

	ctx, ctxCancel := context.WithCancel(context.Background())

	p := &program{
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}

	dialect := generateDialect(cli.HbDisable, cli.StreamreqDisable)

	p.node = &gomavlib.Node{
		Endpoints: endpointConfs,
		Dialect:   dialect,
		OutVersion: func() gomavlib.Version {
			if cli.HbVersion == 2 {
				return gomavlib.V2
			}
			return gomavlib.V1
		}(),
		OutSystemID:            byte(cli.HbSystemid),
		OutComponentID:         byte(cli.HbComponentid),
		HeartbeatDisable:       cli.HbDisable,
		HeartbeatPeriod:        (time.Duration(cli.HbPeriod) * time.Second),
		StreamRequestEnable:    !cli.StreamreqDisable,
		StreamRequestFrequency: cli.StreamreqFrequency,
		ReadTimeout:            cli.ReadTimeout,
		WriteTimeout:           cli.WriteTimeout,
		IdleTimeout:            cli.IdleTimeout,
	}
	err = p.node.Initialize()
	if err != nil {
		ctxCancel()
		return nil, err
	}

	p.errorMan = &errorman.Manager{
		Ctx:               ctx,
		Wg:                &p.wg,
		PrintSingleErrors: cli.PrintErrors,
	}
	err = p.errorMan.Initialize()
	if err != nil {
		ctxCancel()
		p.wg.Wait()
		p.node.Close()
		return nil, err
	}

	p.messageMan = &messageman.Manager{
		Ctx:              ctx,
		Wg:               &p.wg,
		StreamReqDisable: cli.StreamreqDisable,
		Node:             p.node,
	}
	err = p.messageMan.Initialize()
	if err != nil {
		ctxCancel()
		p.wg.Wait()
		p.node.Close()
		return nil, err
	}

	if cli.Dump {
		p.dumper = &dumper.Dumper{
			Ctx:          ctx,
			Wg:           &p.wg,
			Dialect:      dialect,
			DumpPath:     cli.DumpPath,
			DumpDuration: cli.DumpDuration,
		}
		err = p.dumper.Initialize()
		if err != nil {
			ctxCancel()
			p.wg.Wait()
			p.node.Close()
			return nil, err
		}
	}

	if cli.Quiet {
		log.SetOutput(io.Discard)
	}

	log.Printf("mavp2p %s", version)
	log.Printf("router started with %d %s",
		len(endpointConfs),
		func() string {
			if len(endpointConfs) == 1 {
				return "endpoint"
			}
			return "endpoints"
		}())

	p.wg.Add(1)
	go p.run()

	return p, nil
}

func (p *program) close() {
	p.ctxCancel()
	p.wg.Wait()
}

func (p *program) wait() {
	p.wg.Wait()
}

func (p *program) run() {
	defer p.wg.Done()

	defer p.node.Close()

	for {
		select {
		case e := <-p.node.Events():
			switch evt := e.(type) {
			case *gomavlib.EventChannelOpen:
				log.Printf("channel opened: %s", evt.Channel)

			case *gomavlib.EventChannelClose:
				log.Printf("channel closed: %s, %s", evt.Channel, evt.Error)
				p.messageMan.ProcessChannelClose(evt)

			case *gomavlib.EventStreamRequested:
				log.Printf("stream requested to chan=%s sid=%d cid=%d", evt.Channel,
					evt.SystemID, evt.ComponentID)

			case *gomavlib.EventFrame:
				if cli.Print {
					log.Printf("%#v, %#v\n", evt.Frame, evt.Message())
				}
				p.messageMan.ProcessFrame(evt)
				if p.dumper != nil {
					p.dumper.ProcessFrame(evt)
				}

			case *gomavlib.EventParseError:
				p.errorMan.ProcessError(evt)
			}

		case <-p.ctx.Done():
			return
		}
	}
}

func main() {
	p, err := newProgram(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
		os.Exit(1)
	}
	defer p.close()

	p.wait()
}
