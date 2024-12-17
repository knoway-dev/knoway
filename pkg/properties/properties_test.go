package properties

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestGetProperty(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	ctx := InitPropertiesContext(req)

	err := setProperty(ctx, "key", "value")
	require.NoError(t, err)

	v, ok := getProperty[string](ctx, "key")
	require.True(t, ok)
	require.Equal(t, "value", v)
}

func TestPropertyParallel(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	ctx := InitPropertiesContext(req)

	var wg sync.WaitGroup

	for range 1000 {
		wg.Add(1)

		go func(ctx context.Context) {
			err := setProperty(ctx, "key", "value")
			assert.NoError(t, err)

			v, ok := getProperty[string](ctx, "key")
			assert.True(t, ok)
			assert.Equal(t, "value", v)

			wg.Done()
		}(ctx)
	}

	wg.Wait()
}
