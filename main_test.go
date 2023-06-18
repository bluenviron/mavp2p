package main

import (
	"testing"
	"time"

	"github.com/bluenviron/gomavlib/v2"
	"github.com/bluenviron/gomavlib/v2/pkg/dialects/common"
	"github.com/stretchr/testify/require"
)

func TestProgram(t *testing.T) {
	p, err := newProgram([]string{"tcps:0.0.0.0:6666"})
	require.NoError(t, err)
	defer p.close()

	pub, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:6666",
			},
		},
		OutVersion:     gomavlib.V2,
		OutSystemID:    4,
		OutComponentID: 5,
		Dialect:        common.Dialect,
	})
	require.NoError(t, err)
	defer pub.Close()

	sub, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:6666",
			},
		},
		OutVersion:      gomavlib.V2,
		OutSystemID:     6,
		OutComponentID:  7,
		HeartbeatPeriod: 100 * time.Millisecond,
		Dialect:         common.Dialect,
	})
	require.NoError(t, err)
	defer sub.Close()

	<-pub.Events()
	evt := <-pub.Events()
	eventFr, ok := evt.(*gomavlib.EventFrame)
	require.Equal(t, true, ok)
	require.Equal(t, &common.MessageHeartbeat{
		Type:           6,
		SystemStatus:   4,
		MavlinkVersion: 3,
	}, eventFr.Frame.GetMessage())

	msg := &common.MessageOdometry{
		TimeUsec: 123456,
		X:        1.2,
		Y:        2.5,
		Z:        3.4,
	}

	pub.WriteMessageAll(msg)

	<-sub.Events()
	evt = <-sub.Events()
	eventFr, ok = evt.(*gomavlib.EventFrame)
	require.Equal(t, true, ok)
	require.Equal(t, msg, eventFr.Frame.GetMessage())
}
