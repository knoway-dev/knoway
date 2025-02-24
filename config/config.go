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

type AuthServer struct {
	URL     string `yaml:"url" json:"url,omitempty" protobuf:"bytes,1,opt,name=url,proto3"`
	Timeout int64  `yaml:"timeout" json:"timeout,omitempty" protobuf:"varint,2,opt,name=timeout,proto3"` // Unit in seconds
}

type StatsServer struct {
	URL     string `yaml:"url" json:"url,omitempty" protobuf:"bytes,2,opt,name=url,proto3"`
	Timeout int64  `yaml:"timeout" json:"timeout,omitempty" protobuf:"varint,2,opt,name=timeout,proto3"` // Unit in seconds
}

type LogConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled,omitempty"`
}

type Log struct {
	AccessLog *LogConfig `yaml:"access_log" json:"access_log,omitempty"`
}

type GatewayConfig struct {
	AuthServer  AuthServer  `yaml:"auth_server" json:"auth_server"`
	StatsServer StatsServer `yaml:"stats_server" json:"stats_server"`
	Log         *Log        `yaml:"log,omitempty" json:"log,omitempty"`
}

type Config struct {
	Debug      bool             `yaml:"debug" json:"debug"`
	Controller ControllerConfig `yaml:"controller" json:"controller"`
	Gateway    GatewayConfig    `yaml:"gateway" json:"gateway"`
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
