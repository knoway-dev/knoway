package config

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/api/filters/v1alpha1"
	"testing"
)

func TestConfigsServer_LoadFromFile(t *testing.T) {
	newAnypb := func(src proto.Message) *anypb.Any {
		res, _ := anypb.New(src)
		return res
	}
	type fields struct {
		filePath string
	}
	tests := []struct {
		name         string
		fields       fields
		wantErr      bool
		wantClusters map[string]*v1alpha4.Cluster
	}{
		{
			name: "Load valid config",
			fields: fields{
				filePath: "testdata/test_config.json",
			},
			wantErr: false,
			wantClusters: map[string]*v1alpha4.Cluster{
				"llmbackend-sample": {
					Name:              "llmbackend-sample",
					LoadBalancePolicy: 1,
					Filters: []*v1alpha4.ClusterFilter{
						{
							Config: newAnypb(&v1alpha1.UsageStatsConfig{
								StatsServer: &v1alpha1.UsageStatsConfig_StatsServer{
									Url: "0.0.0.0:9090",
								},
							}),
						},
					},
					Provider: "openai",
					Created:  1731986181,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewConfigsServer(tt.fields.filePath)
			err := cs.LoadFromFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				// Check if the loaded clusters match the expected clusters using proto diff
				for key, wantCluster := range tt.wantClusters {
					gotCluster, exists := cs.Clusters[key]
					if !exists {
						t.Errorf("Cluster %s not found", key)
						continue
					}
					assert.Equal(t, wantCluster, gotCluster)
				}
			}
		})
	}
}
