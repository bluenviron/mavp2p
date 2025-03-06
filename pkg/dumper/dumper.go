// Package dumper contains the dump manager.
package dumper

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

var timeNow = time.Now

// Dumper is a dump manager.
type Dumper struct {
	Ctx          context.Context
	Wg           *sync.WaitGroup
	Dialect      *dialect.Dialect
	DumpPath     string
	DumpDuration time.Duration

	dialectRW  *dialect.ReadWriter
	file       *os.File
	tlogWriter *tlog.Writer
	started    time.Time

	chEntry chan *tlog.Entry
}

// Initialize initializes a Dumper.
func (m *Dumper) Initialize() error {
	m.dialectRW = &dialect.ReadWriter{Dialect: m.Dialect}
	err := m.dialectRW.Initialize()
	if err != nil {
		return err
	}

	m.chEntry = make(chan *tlog.Entry, queueSize)

	m.Wg.Add(1)
	go m.run()

	return nil
}

func (m *Dumper) run() {
	defer m.Wg.Done()

	defer func() {
		if m.file != nil {
			m.file.Close()
		}
	}()

	for {
		select {
		case entry := <-m.chEntry:
			err := m.handleEntry(entry)
			if err != nil {
				panic(err)
			}

		case <-m.Ctx.Done():
			return
		}
	}
}

func (m *Dumper) handleEntry(entry *tlog.Entry) error {
	if m.file == nil || entry.Time.Sub(m.started) > m.DumpDuration {
		if m.file != nil {
			m.file.Close()
		}

		m.started = entry.Time

		dir := filepath.Dir(m.DumpPath)
		fname := entry.Time.Format(filepath.Base(m.DumpPath))
		fpath := filepath.Join(dir, fname)

		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}

		m.file, err = os.Create(fpath)
		if err != nil {
			return err
		}

		m.tlogWriter = &tlog.Writer{
			ByteWriter: m.file,
			DialectRW:  m.dialectRW,
		}
		err = m.tlogWriter.Initialize()
		if err != nil {
			return err
		}
	}

	return m.tlogWriter.Write(entry)
}

// ProcessFrame processes a EventFrame.
func (m *Dumper) ProcessFrame(evt *gomavlib.EventFrame) {
	select {
	case m.chEntry <- &tlog.Entry{
		Time:  timeNow(),
		Frame: evt.Frame,
	}:
	case <-m.Ctx.Done():
	default:
		log.Printf("WARN: disk is too slow, discarding frame")
	}
}
