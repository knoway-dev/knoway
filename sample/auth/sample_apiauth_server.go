package main

import (
	"context"
	"log"
	"net"

	pb "knoway.dev/api/service/v1alpha1" // 替换为生成的包路径

	"google.golang.org/grpc"
)

var apiKeys = map[string]struct {
	AllowModels []string
	APIKeyID    string
	UserID      string
}{
	"valid_api_key_123": {
		AllowModels: []string{"*/llama", "kebe/*"},
		APIKeyID:    "apikey_123",
		UserID:      "user_001",
	},
	"valid_api_key_456": {
		AllowModels: []string{},
		APIKeyID:    "apikey_456",
		UserID:      "user_002",
	},
	"valid_api_key_all": {
		AllowModels: []string{"*"},
		APIKeyID:    "apikey_all",
		UserID:      "user_003",
	},
}

// AuthServiceServer 实现
type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *AuthServiceServer) APIKeyAuth(ctx context.Context, req *pb.APIKeyAuthRequest) (*pb.APIKeyAuthResponse, error) {
	if keyData, exists := apiKeys[req.ApiKey]; exists {
		return &pb.APIKeyAuthResponse{
			IsValid:     true,
			AllowModels: keyData.AllowModels,
			ApiKeyId:    keyData.APIKeyID,
			UserId:      keyData.UserID,
		}, nil
	}

	return &pb.APIKeyAuthResponse{
		IsValid:     false,
		AllowModels: nil,
		ApiKeyId:    "",
		UserId:      "",
	}, nil
}

func main() {
	// 创建 gRPC 服务器
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &AuthServiceServer{})

	log.Println("Mock AuthService is running on port 50051...")

	// 启动服务，保持运行
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
