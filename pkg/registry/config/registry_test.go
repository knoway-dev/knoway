package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequestFiltersKeys(t *testing.T) {
	checkKeys := func(expectedKeys []string, actualKeys []string) {
		require.NotEmpty(t, actualKeys)

		for _, expectedKey := range expectedKeys {
			found := false

			for _, key := range actualKeys {
				if key == expectedKey {
					found = true
					break
				}
			}

			assert.True(t, found)
		}
	}

	expectedRequestFiltersKeys := []string{
		"type.googleapis.com/knoway.filters.v1alpha1.APIKeyAuthConfig",
	}
	keys := NewRequestFiltersKeys()
	checkKeys(expectedRequestFiltersKeys, keys)

	expectedClustersFiltersKeys := []string{
		"type.googleapis.com/knoway.filters.v1alpha1.UsageStatsConfig",
	}
	cKeys := NewClustersFiltersKeys()
	checkKeys(expectedClustersFiltersKeys, cKeys)
}
