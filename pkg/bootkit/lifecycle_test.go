package bootkit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLifeCycleHook_Start(t *testing.T) {
	t.Parallel()

	l := LifeCycleHook{}

	require.NotPanics(t, func() {
		err := l.Start(context.Background())
		assert.NoError(t, err)
	})

	l = LifeCycleHook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
	}

	require.NotPanics(t, func() {
		err := l.Start(context.Background())
		assert.NoError(t, err)
	})
}

func TestLifeCycleHook_Stop(t *testing.T) {
	t.Parallel()

	l := LifeCycleHook{}

	require.NotPanics(t, func() {
		err := l.Stop(context.Background())
		assert.NoError(t, err)
	})

	l = LifeCycleHook{
		OnStop: func(ctx context.Context) error {
			return nil
		},
	}

	require.NotPanics(t, func() {
		err := l.Stop(context.Background())
		assert.NoError(t, err)
	})
}

func TestLifeCycle_Append(t *testing.T) {
	t.Parallel()

	l := newLifeCycle()

	l.Append(LifeCycleHook{})
	l.Append(LifeCycleHook{})

	assert.Len(t, l.hooks, 2)
}
