package main

import (
	"github.com/gswly/gomavlib"
	"github.com/gswly/gomavlib/dialects/common"
	"log"
)

type streamHandler struct {
	aprsFrequency    int
	requestedStreams map[remoteNode]struct{}
}

func newStreamHandler(aprsDisable bool, aprsFrequency int) (*streamHandler, error) {
	if aprsDisable == true {
		return nil, nil // disable handler
	}

	sh := &streamHandler{
		aprsFrequency:    aprsFrequency,
		requestedStreams: make(map[remoteNode]struct{}),
	}

	return sh, nil
}

func (sh *streamHandler) onEventFrame(node *gomavlib.Node, evt *gomavlib.EventFrame, rnode remoteNode) bool {
	// request streams to ardupilot devices
	// if they are new or not seen in some time
	if hb, ok := evt.Message().(*common.MessageHeartbeat); ok &&
		hb.Autopilot == common.MAV_AUTOPILOT_ARDUPILOTMEGA {
		if _, ok := sh.requestedStreams[rnode]; !ok {
			sh.requestedStreams[rnode] = struct{}{}

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
