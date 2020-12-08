package main

import (
	"github.com/aler9/gomavlib"
	"log"
	"sync"
	"time"
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
		for {
			time.Sleep(5 * time.Second)

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
