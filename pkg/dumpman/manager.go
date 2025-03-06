// Package dumpman contains the dump manager.
package dumpman

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialect"
	"github.com/bluenviron/gomavlib/v3/pkg/tlog"
)

const (
	queueSize = 128
)

type Manager struct {
	Ctx        context.Context
	Wg         *sync.WaitGroup
	Dialect    *dialect.Dialect
	Dump       bool
	DumpFolder string
	DumpFile   string

	dialectRW *dialect.ReadWriter

	chEntry chan *tlog.Entry
}

// Initialize initializes a Manager.
func (m *Manager) Initialize() error {
	if m.Dump {
		dialectRW := &dialect.ReadWriter{Dialect: m.Dialect}
		err := dialectRW.Initialize()
		if err != nil {
			return err
		}

		m.chEntry = make(chan *tlog.Entry, queueSize)
	}

	m.Wg.Add(1)
	go m.run()

	return nil
}

func (m *Manager) run() {
	defer m.Wg.Done()

	if m.Dump {
		err := os.MkdirAll(m.DumpFolder, 0o755)
		if err != nil {
			panic(err)
		}

		f, err := os.Create(filepath.Join(m.DumpFolder, time.Now().Format(m.DumpFile)))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		tlogWriter := &tlog.Writer{
			ByteWriter: f,
			DialectRW:  m.dialectRW,
		}
		err = tlogWriter.Initialize()
		if err != nil {
			panic(err)
		}

		for {
			select {
			case entry := <-m.chEntry:
				err := tlogWriter.Write(entry)
				if err != nil {
					panic(err)
				}

			case <-m.Ctx.Done():
				return
			}
		}
	} else {
		<-m.Ctx.Done()
	}
}

// ProcessFrame processes a EventFrame.
func (m *Manager) ProcessFrame(evt *gomavlib.EventFrame) {
	if !m.Dump {
		return
	}

	select {
	case m.chEntry <- &tlog.Entry{
		Time:  time.Now(),
		Frame: evt.Frame,
	}:
	case <-m.Ctx.Done():
	default:
		log.Printf("WARN: disk is too slow, discarding frame")
	}
}
