package messageman

import (
	"context"
	"sync"
	"testing"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v3/pkg/frame"
	"github.com/bluenviron/gomavlib/v3/pkg/message"
	"github.com/stretchr/testify/require"
)

func TestRouteSingle(t *testing.T) {
	node := &gomavlib.Node{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPServer{
				Address: "127.0.0.1:3345",
			},
		},
		OutVersion:     gomavlib.V1,
		OutSystemID:    22,
		OutComponentID: 13,
		Dialect:        ardupilotmega.Dialect,
	}
	err := node.Initialize()
	require.NoError(t, err)
	defer node.Close()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	m := &Manager{
		Ctx:              ctx,
		Wg:               &wg,
		StreamReqDisable: true,
		Node:             node,
	}
	err = m.Initialize()
	require.NoError(t, err)

	client := &gomavlib.Node{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:3345",
			},
		},
		OutVersion:     gomavlib.V1,
		OutSystemID:    99,
		OutComponentID: 34,
	}
	err = client.Initialize()
	require.NoError(t, err)
	defer client.Close()

	evt := <-node.Events()
	<-client.Events()
	ch := evt.(*gomavlib.EventChannelOpen).Channel

	fr := &frame.V2Frame{
		SequenceNumber: 127,
		SystemID:       99,
		ComponentID:    34,
		Message:        &ardupilotmega.MessageHeartbeat{},
	}
	err = node.FixFrame(fr)
	require.NoError(t, err)

	m.ProcessFrame(&gomavlib.EventFrame{
		Frame:   fr,
		Channel: ch,
	})

	fr = &frame.V2Frame{
		SequenceNumber: 127,
		SystemID:       30,
		ComponentID:    17,
		Message: &ardupilotmega.MessageOsdParamConfig{
			TargetSystem:    99,
			TargetComponent: 34,
		},
	}
	err = node.FixFrame(fr)
	require.NoError(t, err)

	m.ProcessFrame(&gomavlib.EventFrame{
		Frame: fr,
	})

	evt = <-client.Events()
	require.Equal(t, &message.MessageRaw{
		ID: 11033,
		Payload: []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x63, 0x22,
		},
	}, evt.(*gomavlib.EventFrame).Frame.GetMessage())

	cancel()
	wg.Wait()
}

func TestRouteAll(t *testing.T) {
	node := &gomavlib.Node{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPServer{
				Address: "127.0.0.1:3345",
			},
		},
		OutVersion:     gomavlib.V1,
		OutSystemID:    22,
		OutComponentID: 13,
		Dialect:        ardupilotmega.Dialect,
	}
	err := node.Initialize()
	require.NoError(t, err)
	defer node.Close()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	m := &Manager{
		Ctx:              ctx,
		Wg:               &wg,
		StreamReqDisable: true,
		Node:             node,
	}
	err = m.Initialize()
	require.NoError(t, err)

	client := &gomavlib.Node{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:3345",
			},
		},
		OutVersion:     gomavlib.V1,
		OutSystemID:    99,
		OutComponentID: 34,
	}
	err = client.Initialize()
	require.NoError(t, err)
	defer client.Close()

	<-node.Events()
	<-client.Events()

	fr := &frame.V2Frame{
		SequenceNumber: 127,
		SystemID:       30,
		ComponentID:    17,
		Message: &ardupilotmega.MessageOsdParamConfig{
			TargetSystem:    99,
			TargetComponent: 34,
		},
	}
	err = node.FixFrame(fr)
	require.NoError(t, err)

	m.ProcessFrame(&gomavlib.EventFrame{
		Frame: fr,
	})

	evt := <-client.Events()
	require.Equal(t, &message.MessageRaw{
		ID: 11033,
		Payload: []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x63, 0x22,
		},
	}, evt.(*gomavlib.EventFrame).Frame.GetMessage())

	cancel()
	wg.Wait()
}
