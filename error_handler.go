package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3"
)

type errorHandler struct {
	ctx               context.Context
	wg                *sync.WaitGroup
	printSingleErrors bool

	errorCount      int
	errorCountMutex sync.Mutex
}

func newErrorHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	printSingleErrors bool,
) (*errorHandler, error) {
	eh := &errorHandler{
		ctx:               ctx,
		wg:                wg,
		printSingleErrors: printSingleErrors,
	}

	wg.Add(1)
	go eh.run()

	return eh, nil
}

func (eh *errorHandler) run() {
	defer eh.wg.Done()

	if eh.printSingleErrors {
		return
	}

	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			func() {
				eh.errorCountMutex.Lock()
				defer eh.errorCountMutex.Unlock()

				if eh.errorCount > 0 {
					log.Printf("%d errors in the last 5 seconds", eh.errorCount)
					eh.errorCount = 0
				}
			}()

		case <-eh.ctx.Done():
			return
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
