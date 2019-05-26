package main

import (
	"log"
	"time"
	"sync"
	"github.com/gswly/gomavlib"
)

const (
	NODE_INACTIVE_AFTER = 30 * time.Second
)

type nodeHandler struct {
	remoteNodeMutex sync.Mutex
	remoteNodes     map[remoteNode]time.Time
}

func newNodeHandler() (*nodeHandler, error) {
	nh := &nodeHandler{
		remoteNodes: make(map[remoteNode]time.Time),
	}

	return nh, nil
}

func (nh *nodeHandler) run() {
	// delete remote nodes after a period of inactivity
	for {
		time.Sleep(10 * time.Second)

		func() {
			nh.remoteNodeMutex.Lock()
			defer nh.remoteNodeMutex.Unlock()

			for rnode, t := range nh.remoteNodes {
				if time.Since(t) >= NODE_INACTIVE_AFTER {
					log.Printf("node disappeared: %s", rnode)
					delete(nh.remoteNodes, rnode)
				}
			}
		}()
	}
}

func (nh *nodeHandler) onEventFrame(evt *gomavlib.EventFrame) remoteNode {
	// build remoteNode
	rnode := remoteNode{
		Channel:     evt.Channel,
		SystemId:    evt.SystemId(),
		ComponentId: evt.ComponentId(),
	}

	nh.remoteNodeMutex.Lock()
	defer nh.remoteNodeMutex.Unlock()

	// new remote node
	if _, ok := nh.remoteNodes[rnode]; !ok {
		log.Printf("node appeared: %s", rnode)
	}

	// always update time
	nh.remoteNodes[rnode] = time.Now()

	return rnode
}

func (nh *nodeHandler) onEventChannelClose(evt *gomavlib.EventChannelClose) {
	nh.remoteNodeMutex.Lock()
	defer nh.remoteNodeMutex.Unlock()

	// delete remote nodes associated to channel
	for rnode := range nh.remoteNodes {
		if rnode.Channel == evt.Channel {
			delete(nh.remoteNodes, rnode)
			log.Printf("node disappeared: %s", rnode)
		}
	}
}
