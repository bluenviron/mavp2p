package main

import (
	"testing"
	"time"

	"github.com/bluenviron/gomavlib/v2"
	"github.com/bluenviron/gomavlib/v2/pkg/dialects/common"
	"github.com/stretchr/testify/require"
)

func TestGeneric(t *testing.T) {
	p, err := newProgram([]string{"--print", "tcps:0.0.0.0:6666"})
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

func TestDirect(t *testing.T) {
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

	sub1, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:6666",
			},
		},
		OutVersion:     gomavlib.V2,
		OutSystemID:    6,
		OutComponentID: 7,
		Dialect:        common.Dialect,
	})
	require.NoError(t, err)
	defer sub1.Close()

	sub2, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:6666",
			},
		},
		OutVersion:     gomavlib.V2,
		OutSystemID:    8,
		OutComponentID: 9,
		Dialect:        common.Dialect,
	})
	require.NoError(t, err)
	defer sub2.Close()

	<-pub.Events()
	<-sub1.Events()
	<-sub2.Events()

	sub1.WriteMessageAll(&common.MessageHeartbeat{
		Type:           common.MAV_TYPE_GCS,
		SystemStatus:   4,
		MavlinkVersion: 3,
	})

	sub2.WriteMessageAll(&common.MessageHeartbeat{
		Type:           common.MAV_TYPE_GCS,
		SystemStatus:   4,
		MavlinkVersion: 3,
	})

	for i := 0; i < 2; i++ {
		evt := <-pub.Events()
		eventFr, ok := evt.(*gomavlib.EventFrame)
		require.Equal(t, true, ok)
		require.Equal(t, &common.MessageHeartbeat{
			Type:           6,
			SystemStatus:   4,
			MavlinkVersion: 3,
		}, eventFr.Frame.GetMessage())
	}

	msg := &common.MessageCommandLong{
		TargetSystem:    6,
		TargetComponent: 7,
		Command:         common.MAV_CMD_NAV_FOLLOW,
	}

	pub.WriteMessageAll(msg)

	<-sub1.Events()
	evt := <-sub1.Events()
	eventFr, ok := evt.(*gomavlib.EventFrame)
	require.Equal(t, true, ok)
	require.Equal(t, msg, eventFr.Frame.GetMessage())

	<-sub2.Events()
	select {
	case <-sub2.Events():
		t.Errorf("should not happen")
	case <-time.After(100 * time.Millisecond):
	}
}
