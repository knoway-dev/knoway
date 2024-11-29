package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"

	"knoway.dev/api/service/v1alpha1"
)

// APIKeyAuthResponse 结构体定义与之前相同
type APIKeyAuthResponse struct {
	IsValid     bool     `yaml:"is_valid"`
	AllowModels []string `yaml:"allow_models"`
	APIKeyID    string   `yaml:"api_key_id"`
	UserID      string   `yaml:"user_id"`
}

type APIKeyAuthServer struct {
	v1alpha1.UnimplementedAuthServiceServer
	ValidAPIKeys map[string]*APIKeyAuthResponse // 存储从 YAML 加载的 API Key 信息
}

// 从 YAML 文件加载 API Keys
func loadAPIKeysFromYAML(filePath string) (map[string]*APIKeyAuthResponse, error) {
	var apiKeyConfig struct {
		APIKeys []struct {
			APIKey      string   `yaml:"api_key"`
			IsValid     bool     `yaml:"is_valid"`
			AllowModels []string `yaml:"allow_models"`
			APIKeyID    string   `yaml:"api_key_id"`
			UserID      string   `yaml:"user_id"`
		} `yaml:"api_keys"`
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	// 解析 YAML 数据
	if err := yaml.Unmarshal(data, &apiKeyConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML data: %w", err)
	}

	// 将解析的数据转换为 map 结构
	validAPIKeys := make(map[string]*APIKeyAuthResponse)
	for _, apiKey := range apiKeyConfig.APIKeys {
		validAPIKeys[apiKey.APIKey] = &APIKeyAuthResponse{
			IsValid:     apiKey.IsValid,
			AllowModels: apiKey.AllowModels,
			APIKeyID:    apiKey.APIKeyID,
			UserID:      apiKey.UserID,
		}
	}

	return validAPIKeys, nil
}

// APIKeyAuth 处理 API Key 验证请求
func (s *APIKeyAuthServer) APIKeyAuth(ctx context.Context, req *v1alpha1.APIKeyAuthRequest) (*v1alpha1.APIKeyAuthResponse, error) {
	// 从 YAML 加载的 API Keys
	if res, exists := s.ValidAPIKeys[req.GetApiKey()]; exists {
		// 返回相应的认证结果
		return &v1alpha1.APIKeyAuthResponse{
			IsValid:     res.IsValid,
			AllowModels: res.AllowModels,
			ApiKeyId:    res.APIKeyID,
			UserId:      res.UserID,
		}, nil
	}

	// 如果无效的 API Key 返回错误响应
	return &v1alpha1.APIKeyAuthResponse{
		IsValid:     false,
		AllowModels: []string{},
	}, nil
}

func main() {
	// 从 YAML 文件加载 API Key 配置
	validAPIKeys, err := loadAPIKeysFromYAML("samples/api-key-server/config.yaml")
	if err != nil {
		log.Fatalf("Error loading API keys from YAML: %v", err)
	}

	// 创建 gRPC 服务器实例
	server := grpc.NewServer()

	// 创建 APIKeyAuthServer 实例并注册 API Keys
	authServer := &APIKeyAuthServer{
		ValidAPIKeys: validAPIKeys,
	}

	// 注册 AuthService 服务
	v1alpha1.RegisterAuthServiceServer(server, authServer)

	// 监听指定端口
	listener, err := net.Listen("tcp", ":50051") //nolint:gosec
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 启动 gRPC 服务器
	slog.Info("Starting APIKeyAuthServer on port 50051...")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
