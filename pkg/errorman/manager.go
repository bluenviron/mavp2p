// Package errorman contains the error manager.
package errorman

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3"
)

var printInterval = 5 * time.Second

// Manager is a error manager.
type Manager struct {
	Ctx               context.Context
	Wg                *sync.WaitGroup
	PrintSingleErrors bool

	errorCount      int
	errorCountMutex sync.Mutex
}

// Initialize initializes a Manager.
func (m *Manager) Initialize() error {
	m.Wg.Add(1)
	go m.run()

	return nil
}

func (m *Manager) run() {
	defer m.Wg.Done()

	if m.PrintSingleErrors {
		return
	}

	t := time.NewTicker(printInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			func() {
				m.errorCountMutex.Lock()
				defer m.errorCountMutex.Unlock()

				if m.errorCount > 0 {
					log.Printf("%d errors in the last %s", m.errorCount, printInterval)
					m.errorCount = 0
				}
			}()

		case <-m.Ctx.Done():
			return
		}
	}
}

// ProcessError processes a EventParseError.
func (m *Manager) ProcessError(evt *gomavlib.EventParseError) {
	if m.PrintSingleErrors {
		log.Printf("ERR: %s", evt.Error)
		return
	}

	m.errorCountMutex.Lock()
	defer m.errorCountMutex.Unlock()
	m.errorCount++
}
