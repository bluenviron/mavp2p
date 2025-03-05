package errorman

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/stretchr/testify/require"
)

func TestManager(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup

		m := &Manager{
			Ctx:               ctx,
			Wg:                &wg,
			PrintSingleErrors: true,
		}
		err := m.Initialize()
		require.NoError(t, err)

		m.ProcessError(&gomavlib.EventParseError{
			Error: fmt.Errorf("testing"),
		})

		cancel()
		wg.Wait()
	})

	t.Run("grouped", func(t *testing.T) {
		printInterval = 50 * time.Millisecond

		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup

		m := &Manager{
			Ctx:               ctx,
			Wg:                &wg,
			PrintSingleErrors: false,
		}
		err := m.Initialize()
		require.NoError(t, err)

		m.ProcessError(&gomavlib.EventParseError{
			Error: fmt.Errorf("testing"),
		})

		time.Sleep(100 * time.Millisecond)

		cancel()
		wg.Wait()
	})
}
