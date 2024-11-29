package bootkit

import (
	"context"
	"sync"
)

var _ lifeCycler = LifeCycleHook{}

type lifeCycler interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type LifeCycleHook struct {
	OnStart func(ctx context.Context) error
	OnStop  func(ctx context.Context) error
}

func (l LifeCycleHook) Start(ctx context.Context) error {
	if l.OnStart == nil {
		return nil
	}

	return l.OnStart(ctx)
}

func (l LifeCycleHook) Stop(ctx context.Context) error {
	if l.OnStop == nil {
		return nil
	}

	return l.OnStop(ctx)
}

type LifeCycle interface {
	Append(hook LifeCycleHook)
}

type lifeCycle struct {
	hooks []lifeCycler

	mutex sync.Mutex
}

func (l *lifeCycle) Append(hook LifeCycleHook) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.hooks = append(l.hooks, hook)
}

type EmptyLifeCycle struct{}

func (*EmptyLifeCycle) Append(LifeCycleHook) {}

func newLifeCycle() *lifeCycle {
	return &lifeCycle{
		hooks: make([]lifeCycler, 0),
	}
}

func NewEmptyLifeCycle() LifeCycle {
	return &EmptyLifeCycle{}
}
