package properties

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProperty(t *testing.T) {
	ctx := NewPropertiesContext(context.Background())

	err := SetProperty(ctx, "key", "value")
	require.NoError(t, err)

	v, ok := GetProperty[string](ctx, "key")
	require.True(t, ok)
	require.Equal(t, "value", v)
}

func TestPropertyParallel(t *testing.T) {
	ctx := NewPropertiesContext(context.Background())

	var wg sync.WaitGroup

	for range 1000 {
		wg.Add(1)

		go func(ctx context.Context) {
			err := SetProperty(ctx, "key", "value")
			assert.NoError(t, err)

			v, ok := GetProperty[string](ctx, "key")
			assert.True(t, ok)
			assert.Equal(t, "value", v)

			wg.Done()
		}(ctx)
	}

	wg.Wait()
}
