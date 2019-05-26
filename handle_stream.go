package main

import (
	"github.com/gswly/gomavlib"
	"github.com/gswly/gomavlib/dialects/common"
	"log"
	"time"
)

const (
	STREAM_REQUEST_AGAIN_AFTER_INACTIVITY = 30 * time.Second
)

type streamHandler struct {
	aprsFrequency int
	// we can't use nodeHandler's remoteNodes
	// since we specifically track heartbeats, not generic frames
	lastHeartbeats map[remoteNode]time.Time
}

func newStreamHandler(aprsDisable bool, aprsFrequency int) (*streamHandler, error) {
	if aprsDisable == true {
		return nil, nil // disable handler
	}

	sh := &streamHandler{
		aprsFrequency:  aprsFrequency,
		lastHeartbeats: make(map[remoteNode]time.Time),
	}

	return sh, nil
}

func (sh *streamHandler) onEventFrame(node *gomavlib.Node, evt *gomavlib.EventFrame, rnode remoteNode) bool {
	// node is an ardupilot device
	if hb, ok := evt.Message().(*common.MessageHeartbeat); ok &&
		hb.Autopilot == common.MAV_AUTOPILOT_ARDUPILOTMEGA {
		now := time.Now()

		// request streams if node is new or not seen in some time
		request := false
		if _, ok := sh.lastHeartbeats[rnode]; !ok {
			sh.lastHeartbeats[rnode] = time.Now()
			request = true

		} else {
			if now.Sub(sh.lastHeartbeats[rnode]) >= STREAM_REQUEST_AGAIN_AFTER_INACTIVITY {
				request = true
			}

			// always update last seen
			sh.lastHeartbeats[rnode] = now
		}

		if request == true {
			log.Printf("requesting streams to %s", rnode)

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
					ReqMessageRate:  uint16(sh.aprsFrequency),
					StartStop:       1,
				})
			}
		}
	}

	// stop stream requests from ground stations
	if _, ok := evt.Message().(*common.MessageRequestDataStream); ok {
		return true
	}

	return false
}
