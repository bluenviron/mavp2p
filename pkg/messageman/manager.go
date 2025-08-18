// Package messageman contains the message manager.
package messageman

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"github.com/bluenviron/gomavlib/v3/pkg/message"
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
	SnifferSysid     int

	remoteNodeMutex sync.Mutex
	remoteNodes     map[remoteNodeKey]time.Time
}

// Initialize initializes a Manager.
func (m *Manager) Initialize() error {
	m.remoteNodes = make(map[remoteNodeKey]time.Time)

	m.Wg.Add(1)
	go m.run()

	// print log for sniff mode
	if m.SnifferSysid != 0 {
		log.Printf("sniff mode enabled. route all packet to system id %d", m.SnifferSysid)
	}

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

func (m *Manager) findNodeBySystemID(systemID byte) *remoteNodeKey {
	for key := range m.remoteNodes {
		if key.systemID == systemID {
			return &key
		}
	}
	return nil
}

func (m *Manager) findNodeBySystemAndComponentID(systemID byte, componentID byte) *remoteNodeKey {
	for key := range m.remoteNodes {
		if key.systemID == systemID && key.componentID == componentID {
			return &key
		}
	}
	return nil
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
		var key *remoteNodeKey
		if componentID == 0 {
			key = m.findNodeBySystemID(systemID)
		} else {
			key = m.findNodeBySystemAndComponentID(systemID, componentID)
		}

		if key != nil {
			if key.channel == evt.Channel {
				log.Printf("Warning: channel %s attempted to send message to itself, discarding", key.channel)
			} else {
				m.Node.WriteFrameTo(key.channel, evt.Frame) //nolint:errcheck
			}
		} else {
			log.Printf(
				"Warning: received message addressed to unexistent node with systemID=%d and componentID=%d",
				systemID, componentID)
		}

		// if sniff mode enabled, route packet to sniff system
		if m.SnifferSysid != 0 {
			var key_sniff *remoteNodeKey
			key_sniff = m.findNodeBySystemID(byte(m.SnifferSysid))

			if key_sniff != nil {
				if key_sniff.channel == evt.Channel {
					log.Printf("Warning: channel %s attempted to send message to itself, discarding", key_sniff.channel)
				} else {
					m.Node.WriteFrameTo(key_sniff.channel, evt.Frame) //nolint:errcheck
				}
			} else {
				log.Printf("Warning: Sniff System %d is unexistent node.", m.SnifferSysid)
			}
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
