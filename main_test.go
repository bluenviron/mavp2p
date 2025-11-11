package main

import (
	"testing"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"github.com/stretchr/testify/require"
)

func TestBroadcast(t *testing.T) {
	p, err := newProgram([]string{"tcps:0.0.0.0:6666"})
	require.NoError(t, err)
	defer p.close()

	pub := &gomavlib.Node{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:6666",
			},
		},
		OutVersion:       gomavlib.V2,
		OutSystemID:      4,
		OutComponentID:   5,
		Dialect:          common.Dialect,
		HeartbeatDisable: true,
	}
	err = pub.Initialize()
	require.NoError(t, err)
	defer pub.Close()

	sub := &gomavlib.Node{
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
	}
	err = sub.Initialize()
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

	err = pub.WriteMessageAll(msg)
	require.NoError(t, err)

	<-sub.Events()
	evt = <-sub.Events()
	eventFr, ok = evt.(*gomavlib.EventFrame)
	require.Equal(t, true, ok)
	require.Equal(t, msg, eventFr.Frame.GetMessage())
}

func TestTarget(t *testing.T) {
	p, err := newProgram([]string{"tcps:0.0.0.0:6666"})
	require.NoError(t, err)
	defer p.close()

	pub := &gomavlib.Node{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:6666",
			},
		},
		OutVersion:       gomavlib.V2,
		OutSystemID:      4,
		OutComponentID:   5,
		Dialect:          common.Dialect,
		HeartbeatDisable: true,
	}
	err = pub.Initialize()
	require.NoError(t, err)
	defer pub.Close()

	sub1 := &gomavlib.Node{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:6666",
			},
		},
		OutVersion:       gomavlib.V2,
		OutSystemID:      6,
		OutComponentID:   7,
		Dialect:          common.Dialect,
		HeartbeatDisable: true,
	}
	err = sub1.Initialize()
	require.NoError(t, err)
	defer sub1.Close()

	sub2 := &gomavlib.Node{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:6666",
			},
		},
		OutVersion:       gomavlib.V2,
		OutSystemID:      8,
		OutComponentID:   9,
		Dialect:          common.Dialect,
		HeartbeatDisable: true,
	}
	err = sub2.Initialize()
	require.NoError(t, err)
	defer sub2.Close()

	<-pub.Events()
	<-sub1.Events()
	<-sub2.Events()

	err = sub1.WriteMessageAll(&common.MessageHeartbeat{
		Type:           common.MAV_TYPE_GCS,
		SystemStatus:   4,
		MavlinkVersion: 3,
	})
	require.NoError(t, err)

	err = sub2.WriteMessageAll(&common.MessageHeartbeat{
		Type:           common.MAV_TYPE_GCS,
		SystemStatus:   4,
		MavlinkVersion: 3,
	})
	require.NoError(t, err)

	for range 2 {
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

	err = pub.WriteMessageAll(msg)
	require.NoError(t, err)

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

func TestTargetNotFound(t *testing.T) {
	p, err := newProgram([]string{"tcps:0.0.0.0:6666"})
	require.NoError(t, err)
	defer p.close()

	pub := &gomavlib.Node{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:6666",
			},
		},
		OutVersion:       gomavlib.V2,
		OutSystemID:      4,
		OutComponentID:   5,
		Dialect:          common.Dialect,
		HeartbeatDisable: true,
	}
	err = pub.Initialize()
	require.NoError(t, err)
	defer pub.Close()

	sub := &gomavlib.Node{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTCPClient{
				Address: "127.0.0.1:6666",
			},
		},
		OutVersion:       gomavlib.V2,
		OutSystemID:      8,
		OutComponentID:   9,
		Dialect:          common.Dialect,
		HeartbeatDisable: true,
	}
	err = sub.Initialize()
	require.NoError(t, err)
	defer sub.Close()

	<-pub.Events()
	<-sub.Events()

	err = sub.WriteMessageAll(&common.MessageHeartbeat{
		Type:           common.MAV_TYPE_GCS,
		SystemStatus:   4,
		MavlinkVersion: 3,
	})
	require.NoError(t, err)

	evt := <-pub.Events()
	eventFr, ok := evt.(*gomavlib.EventFrame)
	require.Equal(t, true, ok)
	require.Equal(t, &common.MessageHeartbeat{
		Type:           6,
		SystemStatus:   4,
		MavlinkVersion: 3,
	}, eventFr.Frame.GetMessage())

	msg := &common.MessageCommandLong{
		TargetSystem:    6,
		TargetComponent: 7,
		Command:         common.MAV_CMD_NAV_FOLLOW,
	}

	err = pub.WriteMessageAll(msg)
	require.NoError(t, err)

	evt = <-sub.Events()
	eventFr, ok = evt.(*gomavlib.EventFrame)
	require.Equal(t, true, ok)
	require.Equal(t, msg, eventFr.Frame.GetMessage())
}
