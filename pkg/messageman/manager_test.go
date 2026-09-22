package messageman_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/bluenviron/gomavlib/v4"
	"github.com/bluenviron/gomavlib/v4/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v4/pkg/frame"
	"github.com/bluenviron/gomavlib/v4/pkg/message"
	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mavp2p/pkg/messageman"
)

// routedMessageID is the ID of MessageOsdParamConfig, used to tell routed
// frames apart from heartbeats. Payload fidelity is covered by TestRouteSingle.
const routedMessageID = 11033

// openChannels waits for count channels to open on n.
func openChannels(t *testing.T, n *gomavlib.Node, count int) []*gomavlib.Channel {
	t.Helper()

	var channels []*gomavlib.Channel
	timeout := time.After(5 * time.Second)

	for len(channels) < count {
		select {
		case evt := <-n.Events():
			if op, ok := evt.(*gomavlib.EventChannelOpen); ok {
				channels = append(channels, op.Channel)
			}

		case <-timeout:
			t.Fatalf("timed out waiting for %d channels", count)
		}
	}

	return channels
}

// requireRouted asserts that n receives the routed message.
func requireRouted(t *testing.T, n *gomavlib.Node) {
	t.Helper()

	timeout := time.After(5 * time.Second)

	for {
		select {
		case evt := <-n.Events():
			if fr, ok := evt.(*gomavlib.EventFrame); ok {
				if msg, ok2 := fr.Frame.GetMessage().(*message.MessageRaw); ok2 && msg.ID == routedMessageID {
					return
				}
			}

		case <-timeout:
			t.Fatal("timed out waiting for routed message")
		}
	}
}

// targetedFrame returns a frame addressed to system 99, component targetComponent.
//
// FixFrame encodes the message into a MessageRaw, which would hide the
// TargetSystem/TargetComponent fields that getTarget reads by reflection and
// send the frame down the broadcast path instead. The typed message is put back
// afterwards, which leaves the checksum valid and the frame in the same shape as
// one decoded from an endpoint.
func targetedFrame(t *testing.T, n *gomavlib.Node, targetComponent byte) *frame.V2Frame {
	t.Helper()

	msg := &ardupilotmega.MessageOsdParamConfig{
		TargetSystem:    99,
		TargetComponent: targetComponent,
	}

	fr := &frame.V2Frame{
		SequenceNumber: 127,
		SystemID:       30,
		ComponentID:    17,
		Message:        msg,
	}
	err := n.FixFrame(fr)
	require.NoError(t, err)
	fr.Message = msg

	return fr
}

// announce registers a node on a channel.
func announce(m *messageman.Manager, n *gomavlib.Node, ch *gomavlib.Channel, systemID byte, componentID byte) error {
	fr := &frame.V2Frame{
		SequenceNumber: 1,
		SystemID:       systemID,
		ComponentID:    componentID,
		Message:        &ardupilotmega.MessageHeartbeat{},
	}
	if err := n.FixFrame(fr); err != nil {
		return err
	}

	m.ProcessFrame(&gomavlib.EventFrame{
		Frame:   fr,
		Channel: ch,
	})

	return nil
}

func TestRouteSingle(t *testing.T) {
	node := &gomavlib.Node{
		Endpoints: []gomavlib.Endpoint{
			&gomavlib.EndpointTCPServer{
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

	m := &messageman.Manager{
		Ctx:              ctx,
		Wg:               &wg,
		StreamReqDisable: true,
		Node:             node,
	}
	err = m.Initialize()
	require.NoError(t, err)

	client := &gomavlib.Node{
		Endpoints: []gomavlib.Endpoint{
			&gomavlib.EndpointTCPClient{
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
		Endpoints: []gomavlib.Endpoint{
			&gomavlib.EndpointTCPServer{
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

	m := &messageman.Manager{
		Ctx:              ctx,
		Wg:               &wg,
		StreamReqDisable: true,
		Node:             node,
	}
	err = m.Initialize()
	require.NoError(t, err)

	client := &gomavlib.Node{
		Endpoints: []gomavlib.Endpoint{
			&gomavlib.EndpointTCPClient{
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

// countRouted counts the copies of the routed message that n receives within window.
func countRouted(t *testing.T, n *gomavlib.Node, window time.Duration) int {
	t.Helper()

	count := 0
	deadline := time.After(window)

	for {
		select {
		case evt := <-n.Events():
			if fr, ok := evt.(*gomavlib.EventFrame); ok {
				if msg, ok2 := fr.Frame.GetMessage().(*message.MessageRaw); ok2 && msg.ID == routedMessageID {
					count++
				}
			}

		case <-deadline:
			return count
		}
	}
}

func TestRouteMultipleChannels(t *testing.T) {
	node := &gomavlib.Node{
		Endpoints: []gomavlib.Endpoint{
			&gomavlib.EndpointTCPServer{
				Address: "127.0.0.1:3346",
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

	m := &messageman.Manager{
		Ctx:              ctx,
		Wg:               &wg,
		StreamReqDisable: true,
		Node:             node,
	}
	err = m.Initialize()
	require.NoError(t, err)

	clients := make([]*gomavlib.Node, 2)
	for i := range clients {
		clients[i] = &gomavlib.Node{
			Endpoints: []gomavlib.Endpoint{
				&gomavlib.EndpointTCPClient{
					Address: "127.0.0.1:3346",
				},
			},
			OutVersion:     gomavlib.V1,
			OutSystemID:    99,
			OutComponentID: 34,
		}
		err = clients[i].Initialize()
		require.NoError(t, err)
		defer clients[i].Close() //nolint:gocritic
	}

	channels := openChannels(t, node, 2)

	// both channels claim the same system and component
	for _, ch := range channels {
		err = announce(m, node, ch, 99, 34)
		require.NoError(t, err)
	}

	m.ProcessFrame(&gomavlib.EventFrame{
		Frame: targetedFrame(t, node, 34),
	})

	for _, client := range clients {
		requireRouted(t, client)
	}

	cancel()
	wg.Wait()
}

func TestRouteMultipleChannelsComponentZero(t *testing.T) {
	node := &gomavlib.Node{
		Endpoints: []gomavlib.Endpoint{
			&gomavlib.EndpointTCPServer{
				Address: "127.0.0.1:3347",
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

	m := &messageman.Manager{
		Ctx:              ctx,
		Wg:               &wg,
		StreamReqDisable: true,
		Node:             node,
	}
	err = m.Initialize()
	require.NoError(t, err)

	clients := make([]*gomavlib.Node, 2)
	for i := range clients {
		clients[i] = &gomavlib.Node{
			Endpoints: []gomavlib.Endpoint{
				&gomavlib.EndpointTCPClient{
					Address: "127.0.0.1:3347",
				},
			},
			OutVersion:     gomavlib.V1,
			OutSystemID:    99,
			OutComponentID: byte(34 + i),
		}
		err = clients[i].Initialize()
		require.NoError(t, err)
		defer clients[i].Close() //nolint:gocritic
	}

	channels := openChannels(t, node, 2)

	// same system, different components, one per channel
	for i, ch := range channels {
		err = announce(m, node, ch, 99, byte(34+i))
		require.NoError(t, err)
	}

	// component 0 addresses every component of the system
	m.ProcessFrame(&gomavlib.EventFrame{
		Frame: targetedFrame(t, node, 0),
	})

	for _, client := range clients {
		requireRouted(t, client)
	}

	cancel()
	wg.Wait()
}

func TestRouteSameChannelOnce(t *testing.T) {
	node := &gomavlib.Node{
		Endpoints: []gomavlib.Endpoint{
			&gomavlib.EndpointTCPServer{
				Address: "127.0.0.1:3348",
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

	m := &messageman.Manager{
		Ctx:              ctx,
		Wg:               &wg,
		StreamReqDisable: true,
		Node:             node,
	}
	err = m.Initialize()
	require.NoError(t, err)

	client := &gomavlib.Node{
		Endpoints: []gomavlib.Endpoint{
			&gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:3348",
			},
		},
		OutVersion:     gomavlib.V1,
		OutSystemID:    99,
		OutComponentID: 34,
	}
	err = client.Initialize()
	require.NoError(t, err)
	defer client.Close()

	channels := openChannels(t, node, 1)

	// two components of the same system on a single channel
	err = announce(m, node, channels[0], 99, 34)
	require.NoError(t, err)
	err = announce(m, node, channels[0], 99, 35)
	require.NoError(t, err)

	m.ProcessFrame(&gomavlib.EventFrame{
		Frame: targetedFrame(t, node, 0),
	})

	require.Equal(t, 1, countRouted(t, client, 500*time.Millisecond))

	cancel()
	wg.Wait()
}

func TestRouteToItself(t *testing.T) {
	node := &gomavlib.Node{
		Endpoints: []gomavlib.Endpoint{
			&gomavlib.EndpointTCPServer{
				Address: "127.0.0.1:3349",
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

	m := &messageman.Manager{
		Ctx:              ctx,
		Wg:               &wg,
		StreamReqDisable: true,
		Node:             node,
	}
	err = m.Initialize()
	require.NoError(t, err)

	clients := make([]*gomavlib.Node, 2)
	for i := range clients {
		clients[i] = &gomavlib.Node{
			Endpoints: []gomavlib.Endpoint{
				&gomavlib.EndpointTCPClient{
					Address: "127.0.0.1:3349",
				},
			},
			OutVersion:     gomavlib.V1,
			OutSystemID:    99,
			OutComponentID: 34,
		}
		err = clients[i].Initialize()
		require.NoError(t, err)
		defer clients[i].Close() //nolint:gocritic
	}

	channels := openChannels(t, node, 2)

	err = announce(m, node, channels[0], 99, 34)
	require.NoError(t, err)

	// the only matching channel is the sender: fall back to broadcast, which
	// delivers one copy, to the other channel. Which client owns which channel
	// is not defined, so count across both.
	m.ProcessFrame(&gomavlib.EventFrame{
		Frame:   targetedFrame(t, node, 34),
		Channel: channels[0],
	})

	total := countRouted(t, clients[0], 500*time.Millisecond) +
		countRouted(t, clients[1], 500*time.Millisecond)
	require.Equal(t, 1, total)

	cancel()
	wg.Wait()
}
