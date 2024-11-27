package bootkit

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	bootkit := New(
		StartTimeout(time.Second*2),
		StopTimeout(time.Second*5),
	)

	assert.Equal(t, time.Second*2, bootkit.options.startTimeout)
	assert.Equal(t, time.Second*5, bootkit.options.stopTimeout)
	assert.NotNil(t, bootkit.parallelRun)
	assert.NotNil(t, bootkit.lifeCycle)
	assert.NotNil(t, bootkit.selfCtx)
	assert.NotNil(t, bootkit.selfCancel)
}

func TestBootkit_Add(t *testing.T) {
	t.Parallel()

	bootkit := New()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
				return nil
			})

			wg.Done()
		}()
	}

	wg.Wait()

	assert.Len(t, bootkit.parallelRun, 100)
}

func TestWaitDoneOrContextDone(t *testing.T) {
	t.Parallel()

	t.Run("ImmediatelyWaitGroupDone", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		wg := new(sync.WaitGroup)
		errChan := make(chan error)

		err := waitDoneOrContextDone(ctx, wg, errChan)
		require.NoError(t, err)
	})

	t.Run("OneSecondWaitGroupDone", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		wg := new(sync.WaitGroup)
		errChan := make(chan error)

		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func() {
				time.Sleep(time.Second)
				wg.Done()
			}()
		}

		err := waitDoneOrContextDone(ctx, wg, errChan)
		require.NoError(t, err)
	})

	t.Run("ContextDoneBeforeWaitGroupDone", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		defer cancel()

		wg := new(sync.WaitGroup)
		errChan := make(chan error)

		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func() {
				time.Sleep(time.Second * 2)
				wg.Done()
			}()
		}

		err := waitDoneOrContextDone(ctx, wg, errChan)
		require.Error(t, err)
		assert.ErrorIs(t, context.DeadlineExceeded, err)
	})

	t.Run("ErrorCircuitBreak", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		defer cancel()

		wg := new(sync.WaitGroup)
		errChan := make(chan error)

		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func() {
				if i%2 == 0 {
					time.Sleep(time.Second * 2)
				} else {
					errChan <- errors.New("error")
				}

				wg.Done()
			}()
		}

		err := waitDoneOrContextDone(ctx, wg, errChan)
		require.Error(t, err)
		assert.EqualError(t, err, "error")
	})
}

func TestCallRunnable(t *testing.T) {
	t.Parallel()

	t.Run("TwoSecondsRunnable", func(t *testing.T) {
		t.Parallel()

		lf := newLifeCycle()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		err := callRunnable(ctx, []Runnable{
			func(ctx context.Context, lifeCycle LifeCycle) error {
				time.Sleep(time.Second * 2)

				lifeCycle.Append(LifeCycleHook{
					OnStart: func(ctx context.Context) error {
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return nil
					},
				})

				return nil
			},
			func(ctx context.Context, lifeCycle LifeCycle) error {
				time.Sleep(time.Second * 2)

				lifeCycle.Append(LifeCycleHook{
					OnStart: func(ctx context.Context) error {
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return nil
					},
				})

				return nil
			},
		}, lf)
		require.NoError(t, err)
		assert.Len(t, lf.hooks, 2)
	})

	t.Run("ContextDoneBeforeRunnable-StartTimeout", func(t *testing.T) {
		t.Parallel()

		lf := newLifeCycle()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		defer cancel()

		err := callRunnable(ctx, []Runnable{
			func(ctx context.Context, lifeCycle LifeCycle) error {
				time.Sleep(time.Second * 2)

				return nil
			},
			func(ctx context.Context, lifeCycle LifeCycle) error {
				time.Sleep(time.Second * 2)

				return nil
			},
		}, lf)
		require.Error(t, err)
		assert.ErrorIs(t, context.DeadlineExceeded, err)
	})

	t.Run("ContextDoneBeforeRunnable-EarlyCancel", func(t *testing.T) {
		t.Parallel()

		lf := newLifeCycle()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		time.AfterFunc(time.Second, cancel)

		err := callRunnable(ctx, []Runnable{
			func(ctx context.Context, lifeCycle LifeCycle) error {
				time.Sleep(time.Second * 2)

				return nil
			},
			func(ctx context.Context, lifeCycle LifeCycle) error {
				time.Sleep(time.Second * 2)

				return nil
			},
		}, lf)
		require.Error(t, err)
		assert.ErrorIs(t, context.Canceled, err)
	})

	t.Run("ErrorCircuitBreak", func(t *testing.T) {
		t.Parallel()

		lf := newLifeCycle()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		err := callRunnable(ctx, []Runnable{
			func(ctx context.Context, lifeCycle LifeCycle) error {
				return errors.New("error")
			},
			func(ctx context.Context, lifeCycle LifeCycle) error {
				return nil
			},
		}, lf)
		require.Error(t, err)
		assert.EqualError(t, err, "error")
	})
}

func TestCallStartHooks(t *testing.T) {
	t.Parallel()

	t.Run("TwoSecondsStartHooks", func(t *testing.T) {
		t.Parallel()

		wg := new(sync.WaitGroup)
		errChan := make(chan error)

		startHookRan := make([]bool, 2)
		stopHookRan := make([]bool, 2)
		hooks := []lifeCycler{
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					time.Sleep(time.Second * 2)
					startHookRan[0] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					time.Sleep(time.Second * 2)
					stopHookRan[0] = true
					return nil
				},
			},
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					time.Sleep(time.Second * 2)
					startHookRan[1] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					time.Sleep(time.Second * 2)
					stopHookRan[1] = true
					return nil
				},
			},
		}

		callStartHook(context.TODO(), wg, errChan, hooks)
		wg.Wait()

		assert.Len(t, errChan, 0)
		assert.True(t, startHookRan[0])
		assert.True(t, startHookRan[1])
		assert.False(t, stopHookRan[0])
		assert.False(t, stopHookRan[1])
	})

	t.Run("Errors", func(t *testing.T) {
		t.Parallel()

		wg := new(sync.WaitGroup)

		startHookRan := make([]bool, 2)
		stopHookRan := make([]bool, 2)
		hooks := []lifeCycler{
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startHookRan[0] = true
					return errors.New("error")
				},
				OnStop: func(ctx context.Context) error {
					stopHookRan[0] = true
					return nil
				},
			},
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startHookRan[1] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopHookRan[1] = true
					return nil
				},
			},
		}

		errChan := make(chan error, len(hooks))
		callStartHook(context.TODO(), wg, errChan, hooks)
		wg.Wait()

		err := <-errChan
		require.Error(t, err)
		assert.EqualError(t, err, "error")

		assert.True(t, startHookRan[0])
		assert.True(t, startHookRan[1])
		assert.False(t, stopHookRan[0])
		assert.False(t, stopHookRan[1])
	})
}

func TestCallStopHooks(t *testing.T) {
	t.Parallel()

	t.Run("TwoSecondsStopHooks", func(t *testing.T) {
		t.Parallel()

		startHookRan := make([]bool, 2)
		stopHookRan := make([]bool, 2)
		hooks := []lifeCycler{
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					time.Sleep(time.Second * 2)
					startHookRan[0] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					time.Sleep(time.Second * 2)
					stopHookRan[0] = true
					return nil
				},
			},
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					time.Sleep(time.Second * 2)
					startHookRan[1] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					time.Sleep(time.Second * 2)
					stopHookRan[1] = true
					return nil
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		err := callStopHooks(ctx, hooks)
		require.NoError(t, err)

		assert.False(t, startHookRan[0])
		assert.False(t, startHookRan[1])
		assert.True(t, stopHookRan[0])
		assert.True(t, stopHookRan[1])
	})

	t.Run("Errors", func(t *testing.T) {
		t.Parallel()

		startHookRan := make([]bool, 2)
		stopHookRan := make([]bool, 2)
		hooks := []lifeCycler{
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startHookRan[0] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopHookRan[0] = true
					return errors.New("error")
				},
			},
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startHookRan[1] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopHookRan[1] = false
					return nil
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		err := callStopHooks(ctx, hooks)
		require.Error(t, err)
		assert.EqualError(t, err, "error")

		assert.False(t, startHookRan[0])
		assert.False(t, startHookRan[1])
		assert.True(t, stopHookRan[0])
		assert.False(t, stopHookRan[1])
	})

	t.Run("ContextDoneBeforeStopHook-StopTimeout", func(t *testing.T) {
		t.Parallel()

		startHookRan := make([]bool, 2)
		stopHookRan := make([]bool, 2)
		hooks := []lifeCycler{
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startHookRan[0] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					time.Sleep(time.Second * 3)
					stopHookRan[0] = true
					return nil
				},
			},
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startHookRan[1] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopHookRan[1] = false
					return nil
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		err := callStopHooks(ctx, hooks)
		require.Error(t, err)
		assert.ErrorIs(t, context.DeadlineExceeded, err)

		assert.False(t, startHookRan[0])
		assert.False(t, startHookRan[1])
		assert.False(t, stopHookRan[0])
		assert.False(t, stopHookRan[1])
	})

	t.Run("ContextDoneBeforeStopHook-EarlyCancel", func(t *testing.T) {
		t.Parallel()

		startHookRan := make([]bool, 2)
		stopHookRan := make([]bool, 2)
		hooks := []lifeCycler{
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startHookRan[0] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					time.Sleep(time.Second * 3)
					stopHookRan[0] = true
					return nil
				},
			},
			LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startHookRan[1] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopHookRan[1] = false
					return nil
				},
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(time.Second, cancel)

		err := callStopHooks(ctx, hooks)
		require.Error(t, err)
		assert.ErrorIs(t, context.Canceled, err)

		assert.False(t, startHookRan[0])
		assert.False(t, startHookRan[1])
		assert.False(t, stopHookRan[0])
		assert.False(t, stopHookRan[1])
	})
}

func TestWaitGroupToChan(t *testing.T) {
	t.Parallel()

	wg := new(sync.WaitGroup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	wg.Add(1)

	go func() {
		time.Sleep(time.Second * 2)
		wg.Done()
	}()

	check := func() (bool, error) {
		for {
			select {
			case <-waitGroupToChan(wg):
				return true, nil
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}
	}

	done, err := check()
	require.NoError(t, err)
	assert.True(t, done)
}

func TestBootkit_Start(t *testing.T) {
	t.Parallel()

	t.Run("TwoSecondsStart", func(t *testing.T) {
		t.Parallel()

		bootkit := New()

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			time.Sleep(time.Second * 2)

			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					time.Sleep(time.Second * 3)

					return nil
				},
				OnStop: func(ctx context.Context) error {
					return nil
				},
			})

			return nil
		})

		start := time.Now()

		bootkit.Start()

		elapsed := time.Since(start)

		assert.GreaterOrEqual(t, elapsed.Milliseconds(), (time.Second * 5).Milliseconds())
	})

	t.Run("RunnableError", func(t *testing.T) {
		t.Parallel()

		bootkit := New()

		startCalled := false
		stopCalled := false

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startCalled = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopCalled = true
					return nil
				},
			})

			return errors.New("error")
		})

		bootkit.Start()

		assert.False(t, startCalled)
		assert.True(t, stopCalled)
	})

	t.Run("LaterRunnableError", func(t *testing.T) {
		t.Parallel()

		bootkit := New()

		startCalled := make([]bool, 2)
		stopCalled := make([]bool, 2)

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startCalled[0] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopCalled[0] = true
					return nil
				},
			})

			return nil
		})

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			time.Sleep(time.Second * 2)

			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startCalled[1] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopCalled[1] = true
					return nil
				},
			})

			return errors.New("error")
		})

		bootkit.Start()

		assert.False(t, startCalled[0])
		assert.False(t, startCalled[1])
		assert.True(t, stopCalled[0])
		assert.True(t, stopCalled[1])
	})

	t.Run("RunnableTimeout", func(t *testing.T) {
		t.Parallel()

		bootkit := New(StartTimeout(time.Second))

		startCalled := make([]bool, 2)
		stopCalled := make([]bool, 2)

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startCalled[0] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopCalled[0] = true
					return nil
				},
			})

			return nil
		})

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startCalled[1] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopCalled[1] = true
					return nil
				},
			})

			time.Sleep(time.Second * 2)

			return nil
		})

		bootkit.Start()

		assert.False(t, startCalled[0])
		assert.False(t, startCalled[1])
		assert.True(t, stopCalled[0])
		assert.True(t, stopCalled[1])
	})

	t.Run("StartTimeout", func(t *testing.T) {
		t.Parallel()

		bootkit := New(StartTimeout(time.Second))

		startCalled := make([]bool, 2)
		stopCalled := make([]bool, 2)

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			time.Sleep(time.Second * 5)

			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startCalled[0] = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopCalled[0] = true
					return nil
				},
			})

			return nil
		})

		bootkit.Start()

		assert.False(t, startCalled[0])
		assert.False(t, stopCalled[0])
	})

	t.Run("Errors", func(t *testing.T) {
		t.Parallel()

		bootkit := New()

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			return errors.New("error")
		})

		bootkit.Start()
	})

	t.Run("Errors-StopHookCalled", func(t *testing.T) {
		t.Parallel()

		bootkit := New()

		startCalled := false
		stopCalled := false

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startCalled = true
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopCalled = true
					return nil
				},
			})

			return errors.New("error")
		})

		bootkit.Start()

		assert.False(t, startCalled)
		assert.True(t, stopCalled)
	})

	t.Run("StartHookErrors-StopHookCalled", func(t *testing.T) {
		t.Parallel()

		bootkit := New()

		startCalled := false
		stopCalled := false

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startCalled = true
					return errors.New("error")
				},
				OnStop: func(ctx context.Context) error {
					stopCalled = true
					return nil
				},
			})

			return nil
		})

		bootkit.Start()

		assert.True(t, startCalled)
		assert.True(t, stopCalled)
	})

	t.Run("BothHookErrors-BothHookCalled", func(t *testing.T) {
		t.Parallel()

		bootkit := New()

		startCalled := false
		stopCalled := false

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startCalled = true
					return errors.New("error")
				},
				OnStop: func(ctx context.Context) error {
					stopCalled = true
					return errors.New("error")
				},
			})

			return nil
		})

		bootkit.Start()

		assert.True(t, startCalled)
		assert.True(t, stopCalled)
	})
}

func TestBootkit_Stop(t *testing.T) {
	t.Parallel()

	t.Run("Stop", func(t *testing.T) {
		t.Parallel()

		bootkit := New()

		startCalled := false
		stopCalled := false

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			lifeCycle.Append(LifeCycleHook{
				OnStart: func(ctx context.Context) error {
					startCalled = true
					time.Sleep(time.Second * 2)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					stopCalled = true
					return nil
				},
			})

			return nil
		})

		time.AfterFunc(time.Second, func() {
			err := bootkit.Stop(context.Background())
			assert.NoError(t, err)
		})

		bootkit.Start()

		assert.True(t, startCalled)
		assert.True(t, stopCalled)
	})

	t.Run("Stop-NoStopHook", func(t *testing.T) {
		t.Parallel()

		bootkit := New()

		bootkit.Add(func(ctx context.Context, lifeCycle LifeCycle) error {
			return nil
		})

		time.AfterFunc(time.Second, func() {
			err := bootkit.Stop(context.Background())
			assert.NoError(t, err)
		})

		bootkit.Start()
	})
}
