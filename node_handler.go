package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v2"
	"github.com/bluenviron/gomavlib/v2/pkg/dialects/common"
)

const (
	nodeInactiveAfter = 30 * time.Second
)

type remoteNodeKey struct {
	Channel     *gomavlib.Channel
	SystemID    byte
	ComponentID byte
}

func (i remoteNodeKey) String() string {
	return fmt.Sprintf("chan=%s sid=%d cid=%d", i.Channel, i.SystemID, i.ComponentID)
}

type nodeHandler struct {
	ctx              context.Context
	wg               *sync.WaitGroup
	streamreqDisable bool
	node             *gomavlib.Node

	remoteNodeMutex sync.Mutex
	remoteNodes     map[remoteNodeKey]time.Time
}

func newNodeHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	streamreqDisable bool,
	node *gomavlib.Node,
) (*nodeHandler, error) {
	nh := &nodeHandler{
		ctx:              ctx,
		wg:               wg,
		streamreqDisable: streamreqDisable,
		node:             node,
		remoteNodes:      make(map[remoteNodeKey]time.Time),
	}

	wg.Add(1)
	go nh.run()

	return nh, nil
}

func (nh *nodeHandler) run() {
	defer nh.wg.Done()

	// delete remote nodes after a period of inactivity
	for {
		select {
		case <-time.After(10 * time.Second):
			func() {
				now := time.Now()

				nh.remoteNodeMutex.Lock()
				defer nh.remoteNodeMutex.Unlock()

				for rnode, t := range nh.remoteNodes {
					if now.Sub(t) >= nodeInactiveAfter {
						log.Printf("node disappeared: %s", rnode)
						delete(nh.remoteNodes, rnode)
					}
				}
			}()

		case <-nh.ctx.Done():
			return
		}
	}
}

func (nh *nodeHandler) onEventFrame(evt *gomavlib.EventFrame) {
	key := remoteNodeKey{
		Channel:     evt.Channel,
		SystemID:    evt.SystemID(),
		ComponentID: evt.ComponentID(),
	}

	nh.remoteNodeMutex.Lock()
	defer nh.remoteNodeMutex.Unlock()

	// new remote node
	if _, ok := nh.remoteNodes[key]; !ok {
		log.Printf("node appeared: %s", key)
	}

	// update time
	nh.remoteNodes[key] = time.Now()

	// if automatic stream requests are enabled, block manual stream requests
	if !nh.streamreqDisable {
		if _, ok := evt.Message().(*common.MessageRequestDataStream); ok {
			return
		}
	}

	routeMsg := false // Flag to mark a msg for routing
	targetSystem := uint8(0)
	targetComponent := uint8(0)
	switch msg := evt.Message().(type) {
	case *common.MessageCommandLong:
		routeMsg = true
		targetSystem = msg.TargetSystem
		targetComponent = msg.TargetComponent
	case *common.MessageCommandAck:
		routeMsg = true
		targetSystem = msg.TargetSystem
		targetComponent = msg.TargetComponent
	case *common.MessageCommandInt:
		routeMsg = true
		targetSystem = msg.TargetSystem
		targetComponent = msg.TargetComponent
	}

	if routeMsg {
		if targetSystem > 0 { // Route only if it's non-broadcast command
			for remoteNode := range nh.remoteNodes { // Iterates through connected nodes
				if remoteNode.SystemID == targetSystem {
					if remoteNode.ComponentID == targetComponent ||
						targetComponent < 1 { // Route if compid matches or is a broadcast
						if remoteNode.Channel != evt.Channel { // Prevents Loops
							nh.node.WriteFrameTo(remoteNode.Channel, evt.Frame)
						} else {
							log.Println("Warning: channel ", remoteNode.Channel, " attempted to send to itself, discarding ")
						}
					}
				}
			}
			return
		}
	}

	// route message to every other channel
	nh.node.WriteFrameExcept(evt.Channel, evt.Frame)
}

func (nh *nodeHandler) onEventChannelClose(evt *gomavlib.EventChannelClose) {
	nh.remoteNodeMutex.Lock()
	defer nh.remoteNodeMutex.Unlock()

	// delete remote nodes associated to channel
	for key := range nh.remoteNodes {
		if key.Channel == evt.Channel {
			delete(nh.remoteNodes, key)
			log.Printf("node disappeared: %s", key)
		}
	}
}
