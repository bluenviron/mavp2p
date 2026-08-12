// Package messageman contains the message manager.
package messageman

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v4"
	"github.com/bluenviron/gomavlib/v4/pkg/dialects/common"
	"github.com/bluenviron/gomavlib/v4/pkg/message"
)

const (
	nodeInactiveAfter = 30 * time.Second
)

var zero reflect.Value

func getTarget(msg message.Message) (byte, byte, bool) {
	rv := reflect.ValueOf(msg).Elem()
	ts := rv.FieldByName("TargetSystem")
	tc := rv.FieldByName("TargetComponent")

	if ts != zero && tc != zero {
		return byte(ts.Uint()), byte(tc.Uint()), true
	}

	return 0, 0, false
}

type remoteNodeKey struct {
	channel     *gomavlib.Channel
	systemID    byte
	componentID byte
}

func (i remoteNodeKey) String() string {
	return fmt.Sprintf("chan=%s sid=%d cid=%d", i.channel, i.systemID, i.componentID)
}

// Manager is a message manager.
type Manager struct {
	Ctx              context.Context
	Wg               *sync.WaitGroup
	StreamReqDisable bool
	Node             *gomavlib.Node

	remoteNodeMutex sync.Mutex
	remoteNodes     map[remoteNodeKey]time.Time
}

// Initialize initializes a Manager.
func (m *Manager) Initialize() error {
	m.remoteNodes = make(map[remoteNodeKey]time.Time)

	m.Wg.Add(1)
	go m.run()

	return nil
}

func (m *Manager) run() {
	defer m.Wg.Done()

	// delete remote nodes after a period of inactivity
	for {
		select {
		case <-time.After(10 * time.Second):
			func() {
				now := time.Now()

				m.remoteNodeMutex.Lock()
				defer m.remoteNodeMutex.Unlock()

				for rnode, t := range m.remoteNodes {
					if now.Sub(t) >= nodeInactiveAfter {
						log.Printf("node disappeared: %s", rnode)
						delete(m.remoteNodes, rnode)
					}
				}
			}()

		case <-m.Ctx.Done():
			return
		}
	}
}

// appendChannel adds ch unless already present.
func appendChannel(channels []*gomavlib.Channel, ch *gomavlib.Channel) []*gomavlib.Channel {
	for _, existing := range channels {
		if existing == ch {
			return channels
		}
	}
	return append(channels, ch)
}

// findChannelsBySystemID returns all channels of a given system.
func (m *Manager) findChannelsBySystemID(systemID byte) []*gomavlib.Channel {
	// lock: the cleanup routine deletes from remoteNodes concurrently
	m.remoteNodeMutex.Lock()
	defer m.remoteNodeMutex.Unlock()

	var channels []*gomavlib.Channel
	for key := range m.remoteNodes {
		if key.systemID == systemID {
			channels = appendChannel(channels, key.channel)
		}
	}
	return channels
}

// findChannelsBySystemAndComponentID returns all channels of a given system and component.
func (m *Manager) findChannelsBySystemAndComponentID(systemID byte, componentID byte) []*gomavlib.Channel {
	// lock: the cleanup routine deletes from remoteNodes concurrently
	m.remoteNodeMutex.Lock()
	defer m.remoteNodeMutex.Unlock()

	var channels []*gomavlib.Channel
	for key := range m.remoteNodes {
		if key.systemID == systemID && key.componentID == componentID {
			channels = appendChannel(channels, key.channel)
		}
	}
	return channels
}

// ProcessFrame processes a EventFrame.
func (m *Manager) ProcessFrame(evt *gomavlib.EventFrame) {
	key := remoteNodeKey{
		channel:     evt.Channel,
		systemID:    evt.SystemID(),
		componentID: evt.ComponentID(),
	}

	func() {
		m.remoteNodeMutex.Lock()
		defer m.remoteNodeMutex.Unlock()

		if _, ok := m.remoteNodes[key]; !ok {
			log.Printf("node appeared: %s", key)
		}

		m.remoteNodes[key] = time.Now()
	}()

	// stop stream request messages
	if !m.StreamReqDisable {
		if _, ok := evt.Message().(*common.MessageRequestDataStream); ok {
			return
		}
	}

	// if message has a target, route only to it
	systemID, componentID, hasTarget := getTarget(evt.Message())
	if hasTarget && systemID > 0 {
		var channels []*gomavlib.Channel
		if componentID == 0 {
			channels = m.findChannelsBySystemID(systemID)
		} else {
			channels = m.findChannelsBySystemAndComponentID(systemID, componentID)
		}

		if len(channels) != 0 {
			// a target can be present on multiple channels; route to all of them
			delivered := false
			for _, channel := range channels {
				if channel == evt.Channel {
					continue
				}
				m.Node.WriteFrameTo(channel, evt.Frame) //nolint:errcheck
				delivered = true
			}
			if delivered {
				return
			}
			log.Printf("Warning: channel %s attempted to send message to itself, discarding", evt.Channel)
		} else {
			log.Printf(
				"Warning: received message addressed to unexistent node with systemID=%d and componentID=%d",
				systemID, componentID)
		}
	}

	// otherwise, route message to every channel
	m.Node.WriteFrameExcept(evt.Channel, evt.Frame) //nolint:errcheck
}

// ProcessChannelClose processes a EventChannelClose.
func (m *Manager) ProcessChannelClose(evt *gomavlib.EventChannelClose) {
	m.remoteNodeMutex.Lock()
	defer m.remoteNodeMutex.Unlock()

	// delete remote nodes associated to channel
	for key := range m.remoteNodes {
		if key.channel == evt.Channel {
			delete(m.remoteNodes, key)
			log.Printf("node disappeared: %s", key)
		}
	}
}
