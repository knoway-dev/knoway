package bootkit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBootkit_StopRunningService(t *testing.T) {
	for _, callers := range []int{1, 8} {
		t.Run(fmt.Sprintf("%dCallers", callers), func(t *testing.T) {
			bootkit := New()
			ready := make(chan struct{})
			serviceDone := make(chan struct{})
			startDone := make(chan struct{})
			stopResults := make(chan error, callers)
			var shutdown sync.Once
			stopService := func() { shutdown.Do(func() { close(serviceDone) }) }
			defer stopService()

			bootkit.Add(func(_ context.Context, lifecycle LifeCycle) error {
				lifecycle.Append(LifeCycleHook{
					OnStart: func(context.Context) error {
						close(ready)
						<-serviceDone
						return nil
					},
					OnStop: func(context.Context) error {
						stopService()
						return nil
					},
				})
				return nil
			})

			go func() {
				defer close(startDone)
				bootkit.Start()
			}()
			select {
			case <-ready:
			case <-time.After(time.Second * 5):
				t.Fatal("service did not start")
			}

			for range callers {
				go func() { stopResults <- bootkit.Stop(context.Background()) }()
			}

			stopTimedOut := false
			for range callers {
				select {
				case err := <-stopResults:
					require.NoError(t, err)
				case <-time.After(time.Second):
					stopTimedOut = true
					// Let the old implementation finish so a failing regression leaves
					// no Start or Stop goroutines waiting on the service.
					stopService()
					select {
					case err := <-stopResults:
						require.NoError(t, err)
					case <-time.After(time.Second * 5):
						t.Fatal("Stop did not return after releasing the service")
					}
				}
			}

			select {
			case <-startDone:
			case <-time.After(time.Second * 5):
				t.Fatal("Start did not return after shutdown")
			}
			require.False(t, stopTimedOut, "Stop could not interrupt the running service")
			require.NoError(t, bootkit.Stop(context.Background()))
		})
	}
}
