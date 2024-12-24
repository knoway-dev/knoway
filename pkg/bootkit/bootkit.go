package bootkit

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/samber/lo"

	"knoway.dev/pkg/utils"
)

const (
	DefaultStartTimeout = time.Second * 15
	DefaultStopTimeout  = time.Second * 60
)

type Runnable func(ctx context.Context, lifeCycle LifeCycle) error

type BootKit struct {
	options     *bootkitOptions
	parallelRun []Runnable
	lifeCycle   *lifeCycle

	selfCtx    context.Context
	selfCancel context.CancelFunc

	mutex sync.Mutex
}

func New(options ...Option) *BootKit {
	applyOptions := &bootkitApplyOptions{
		bootkit: &bootkitOptions{
			startTimeout: DefaultStartTimeout,
			stopTimeout:  DefaultStopTimeout,
		},
	}

	for _, opt := range options {
		opt.apply(applyOptions)
	}

	selfCtx, selfCancel := context.WithCancel(context.Background())

	return &BootKit{
		options:     applyOptions.bootkit,
		parallelRun: make([]Runnable, 0),
		lifeCycle:   newLifeCycle(),
		selfCtx:     selfCtx,
		selfCancel:  selfCancel,
	}
}

func (b *BootKit) Add(invokeFn Runnable) *BootKit {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.parallelRun = append(b.parallelRun, invokeFn)

	return b
}

func waitDoneOrContextDone(ctx context.Context, wg *sync.WaitGroup, errChan chan error) error {
	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}

	return nil
}

func callRunnable(ctx context.Context, runnable []Runnable, lifecycle LifeCycle) error {
	wg := sync.WaitGroup{}
	errChan := make(chan error)

	for _, r := range runnable {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := r(ctx, lifecycle)
			if err != nil {
				errChan <- err
			}
		}()
	}

	return waitDoneOrContextDone(ctx, &wg, errChan)
}

func callStartHook(ctx context.Context, startWg *sync.WaitGroup, errChan chan error, hooks []lifeCycler) {
	for _, r := range hooks {
		startWg.Add(1)

		go func() {
			defer startWg.Done()

			err := r.Start(ctx)
			if err != nil {
				errChan <- err
			}
		}()
	}
}

func callStopHooks(ctx context.Context, hooks []lifeCycler) error {
	wg := sync.WaitGroup{}
	errChan := make(chan error)

	for _, hook := range lo.Reverse(utils.Clone(hooks)) {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := hook.Stop(ctx)
			if err != nil {
				errChan <- err
			}
		}()
	}

	return waitDoneOrContextDone(ctx, &wg, errChan)
}

func waitGroupToChan(wg *sync.WaitGroup) <-chan struct{} {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		wg.Wait()
		cancel()
	}()

	return ctx.Done()
}

func (b *BootKit) Start() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	startWg := &sync.WaitGroup{}

	ctx, cancel := context.WithTimeout(b.selfCtx, b.options.startTimeout)
	defer cancel()

	err := callRunnable(ctx, b.parallelRun, b.lifeCycle)
	if err != nil {
		slog.Error("failed to run", "error", err)
		b.mayStop()

		return
	}

	errChan := make(chan error, len(b.lifeCycle.GetHooks()))
	callStartHook(ctx, startWg, errChan, b.lifeCycle.GetHooks())

	go func() {
		sigs := make(chan os.Signal, 2) //nolint:mnd
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		cancelled := false

		for range sigs {
			// Double signal will force exit
			if cancelled {
				fmt.Fprintln(os.Stderr, "received signal, force terminated")
				os.Exit(1)
			}

			b.selfCancel()
			cancelled = true
		}
	}()

	defer b.mayStop()

	for {
		select {
		case err := <-errChan:
			slog.Error("failed to start", "error", err)
			return
		case <-waitGroupToChan(startWg):
			return
		case <-b.selfCtx.Done():
			return
		}
	}
}

func (b *BootKit) stop() error {
	if len(b.lifeCycle.GetHooks()) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.options.stopTimeout)
	defer cancel()

	return callStopHooks(ctx, b.lifeCycle.GetHooks())
}

func (b *BootKit) mayStop() {
	err := b.stop()
	if err != nil {
		slog.Error("failed to stop", "error", err)
	}
}

func (b *BootKit) Stop(ctx context.Context) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.selfCancel()

	return b.stop()
}
