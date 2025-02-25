package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ControllerConfig struct {
	EnableLeaderElection bool `yaml:"enable_leader_election" json:"enable_leader_election"`
	SecureMetrics        bool `yaml:"secure_metrics" json:"secure_metrics"`
	EnableHTTP2          bool `yaml:"enable_http2" json:"enable_http_2"`
}

type Config struct {
	Debug      bool             `yaml:"debug" json:"debug"`
	Controller ControllerConfig `yaml:"controller" json:"controller"`
	// KubeConfig is the path to the kubeconfig file, used for local development, if empty, in-cluster config will be used.
	KubeConfig string `yaml:"kubeConfig" json:"kubeConfig"`

	StaticListeners map[string]map[string]interface{} `yaml:"staticListeners" json:"staticListeners"`
}

// LoadConfig loads the configuration from the specified YAML file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &cfg, nil
}
