package main

import (
	"log"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v2"
)

type errorHandler struct {
	printSingleErrors bool
	errorCount        int
	errorCountMutex   sync.Mutex
}

func newErrorHandler(printSingleErrors bool) (*errorHandler, error) {
	eh := &errorHandler{
		printSingleErrors: printSingleErrors,
	}

	return eh, nil
}

func (eh *errorHandler) run() {
	// print errors in group
	if !eh.printSingleErrors {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()

		for range t.C {
			func() {
				eh.errorCountMutex.Lock()
				defer eh.errorCountMutex.Unlock()

				if eh.errorCount > 0 {
					log.Printf("%d errors in the last 5 seconds", eh.errorCount)
					eh.errorCount = 0
				}
			}()
		}
	}
}

func (eh *errorHandler) onEventError(evt *gomavlib.EventParseError) {
	if eh.printSingleErrors {
		log.Printf("ERR: %s", evt.Error)
		return
	}

	eh.errorCountMutex.Lock()
	defer eh.errorCountMutex.Unlock()
	eh.errorCount++
}
