package config

import (
	"testing"
)

func TestNewRequestFiltersKeys(t *testing.T) {
	checkKeys := func(expectedKeys []string, actualKeys []string) {
		if len(actualKeys) == 0 {
			t.Error("Expected keys to be non-empty")
		}
		for _, expectedKey := range expectedKeys {
			found := false
			for _, key := range actualKeys {
				if key == expectedKey {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected key %q not found in keys", expectedKey)
			}
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
