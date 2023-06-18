package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v2"
)

const (
	nodeInactiveAfter = 30 * time.Second
)

type remoteNode struct {
	Channel     *gomavlib.Channel
	SystemID    byte
	ComponentID byte
}

func (i remoteNode) String() string {
	return fmt.Sprintf("chan=%s sid=%d cid=%d", i.Channel, i.SystemID, i.ComponentID)
}

type nodeHandler struct {
	ctx context.Context
	wg  *sync.WaitGroup

	remoteNodeMutex sync.Mutex
	remoteNodes     map[remoteNode]time.Time
}

func newNodeHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
) (*nodeHandler, error) {
	nh := &nodeHandler{
		ctx:         ctx,
		wg:          wg,
		remoteNodes: make(map[remoteNode]time.Time),
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
	rnode := remoteNode{
		Channel:     evt.Channel,
		SystemID:    evt.SystemID(),
		ComponentID: evt.ComponentID(),
	}

	nh.remoteNodeMutex.Lock()
	defer nh.remoteNodeMutex.Unlock()

	// new remote node
	if _, ok := nh.remoteNodes[rnode]; !ok {
		log.Printf("node appeared: %s", rnode)
	}

	// always update time
	nh.remoteNodes[rnode] = time.Now()
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
