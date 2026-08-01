package dumper_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/bluenviron/gomavlib/v4"
	"github.com/bluenviron/gomavlib/v4/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v4/pkg/frame"
	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mavp2p/pkg/dumper"
)

func TestDumper(t *testing.T) {
	tmpFolder, err := os.MkdirTemp("", "mavp2p-dumper")
	require.NoError(t, err)
	defer os.RemoveAll(tmpFolder)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	d := &dumper.Dumper{
		Ctx:          ctx,
		Wg:           &wg,
		Dialect:      ardupilotmega.Dialect,
		DumpPath:     filepath.Join(tmpFolder, "2006-01-02_15-04-05.000000000.tlog"),
		DumpDuration: 1 * time.Second,
	}
	err = d.Initialize()
	require.NoError(t, err)

	d.ProcessFrame(&gomavlib.EventFrame{
		Frame: &frame.V2Frame{
			SequenceNumber: 123,
			SystemID:       14,
			ComponentID:    15,
			Message:        &ardupilotmega.MessageOsdParamConfig{},
			Checksum:       1234,
		},
	})

	time.Sleep(1100 * time.Millisecond)

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

	entries, err := os.ReadDir(tmpFolder)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	for _, entry := range entries {
		require.Equal(t, ".tlog", filepath.Ext(entry.Name()))
	}
}
