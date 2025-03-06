package dumper

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v3/pkg/frame"
	"github.com/stretchr/testify/require"
)

func TestDumper(t *testing.T) {
	tmpFolder, err := os.MkdirTemp("", "mavp2p-dumper")
	require.NoError(t, err)
	defer os.RemoveAll(tmpFolder)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	d := &Dumper{
		Ctx:          ctx,
		Wg:           &wg,
		Dialect:      ardupilotmega.Dialect,
		DumpPath:     filepath.Join(tmpFolder, "2006-01-02_15-04-05.tlog"),
		DumpDuration: 1 * time.Second,
	}
	err = d.Initialize()
	require.NoError(t, err)

	timeNow = func() time.Time {
		return time.Date(2009, 5, 20, 22, 15, 25, 427000, time.Local)
	}

	d.ProcessFrame(&gomavlib.EventFrame{
		Frame: &frame.V2Frame{
			SequenceNumber: 123,
			SystemID:       14,
			ComponentID:    15,
			Message:        &ardupilotmega.MessageOsdParamConfig{},
			Checksum:       1234,
		},
	})

	timeNow = func() time.Time {
		return time.Date(2009, 5, 20, 22, 16, 25, 427000, time.Local)
	}

	d.ProcessFrame(&gomavlib.EventFrame{
		Frame: &frame.V2Frame{
			SequenceNumber: 123,
			SystemID:       14,
			ComponentID:    15,
			Message:        &ardupilotmega.MessageOsdParamConfig{},
			Checksum:       1234,
		},
	})

	time.Sleep(100 * time.Millisecond)

	cancel()
	wg.Wait()

	_, err = os.Stat(filepath.Join(tmpFolder, "2009-05-20_22-15-25.tlog"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tmpFolder, "2009-05-20_22-16-25.tlog"))
	require.NoError(t, err)
}
