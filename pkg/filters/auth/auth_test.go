package auth

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "knoway.dev/api/service/v1alpha1" // 替换为生成的包路径

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

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
}

// AuthServiceServer 实现
type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *AuthServiceServer) APIKeyAuth(ctx context.Context, req *pb.APIKeyAuthRequest) (*pb.APIKeyAuthResponse, error) {
	if keyData, exists := apiKeys[req.GetApiKey()]; exists {
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

func startTestServer() (*grpc.Server, *bufconn.Listener) {
	listener := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, &AuthServiceServer{})

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	return server, listener
}

func dialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, s string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestAPIKeyAuth(t *testing.T) {
	// 启动测试服务器
	server, listener := startTestServer()
	defer server.Stop()

	// 创建 gRPC 客户端连接
	conn, err := grpc.DialContext( //nolint:staticcheck
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(dialer(listener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	tests := []struct {
		name      string
		apiKey    string
		wantValid bool
	}{
		{"ValidAPIKey1", "valid_api_key_123", true},
		{"ValidAPIKey2", "valid_api_key_456", true},
		{"InvalidAPIKey", "invalid_api_key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.APIKeyAuth(context.Background(), &pb.APIKeyAuthRequest{ApiKey: tt.apiKey})
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, resp.GetIsValid())
		})
	}
}

func TestCanAccessModel(t *testing.T) {
	tests := []struct {
		name         string
		allowModels  []string
		deniedModels []string
		requestModel string
		want         bool
	}{
		{
			name:         "t1",
			allowModels:  []string{"*/*"},
			requestModel: "kebe/model-1",
			want:         true,
		},
		{
			name:         "t2",
			allowModels:  []string{"**"},
			requestModel: "gpt-1",
			want:         true,
		},
		{
			name:         "t3",
			allowModels:  []string{"kebe/*"},
			requestModel: "kebe/model-1",
			want:         true,
		},
		{
			name:         "t4",
			allowModels:  []string{"kebe/*"},
			requestModel: "gpt-1",
			want:         false,
		},
		{
			name:         "t4",
			allowModels:  []string{"kebe/*"},
			requestModel: "nicole/gpt-1",
			want:         false,
		},
		{
			name:         "only public",
			allowModels:  []string{"*"},
			requestModel: "nicole/gpt-1",
			want:         false,
		},
		{
			name:         "* match public",
			allowModels:  []string{"*"},
			requestModel: "gpt-1",
			want:         true,
		},
		{
			name:         "denied priority",
			allowModels:  []string{"*"},
			deniedModels: []string{"gpt-1"},
			requestModel: "gpt-1",
			want:         false,
		},
		{
			name:         "denied priority",
			allowModels:  []string{"*", "public/*", "u-kebe/llama"},
			deniedModels: []string{"*", "public/*"},
			requestModel: "u-kebe/llama",
			want:         true,
		},
		{
			name:         "denied priority",
			allowModels:  []string{"*", "public/*", "u-kebe/llama"},
			deniedModels: []string{"*", "public/*"},
			requestModel: "public/openai",
			want:         false,
		},
		{
			name:         "denied priority with nested layers",
			allowModels:  []string{"*", "public/*", "u-kebe/llama"},
			deniedModels: []string{"*", "public/*"},
			requestModel: "public/openai/gpt-1",
			want:         false,
		},
		{
			name:         "nested layers",
			allowModels:  []string{"public/**"},
			requestModel: "public/openrouter/qwen/qwen-2-7b-instruct",
			want:         true,
		},
		{
			name:         "nested layers not accessible",
			allowModels:  []string{"other-ns/**"},
			requestModel: "public/openrouter/qwen/qwen-2-7b-instruct",
			want:         false,
		},
		{
			name:         "with colon",
			allowModels:  []string{"public/*"},
			requestModel: "public/qwen-2-7b-instruct:32b",
			want:         true,
		},
		{
			name:         "nested layers with allow with colon",
			allowModels:  []string{"public/**"},
			requestModel: "public/openrouter/qwen/qwen-2-7b-instruct:32b",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CanAccessModel(tt.requestModel, tt.allowModels, tt.deniedModels), "CanAccessModel(%v, %v)", tt.allowModels, tt.requestModel)
		})
	}
}
