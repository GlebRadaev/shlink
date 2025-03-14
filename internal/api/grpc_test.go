package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/service/health"
	"github.com/GlebRadaev/shlink/internal/service/url"
	"github.com/GlebRadaev/shlink/pkg/shlink"
)

const bufSize = 1024 * 1024

func TestGRPCServer_Shorten(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis := bufconn.Listen(bufSize)
	defer lis.Close()

	urlService := url.NewMockIURLService(ctrl)
	healthService := health.NewMockIHealthService(ctrl)

	server, err := NewGRPCServer(urlService, healthService)
	require.NoError(t, err)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := shlink.NewShlinkServiceClient(conn)

	testCases := []struct {
		name        string
		req         *shlink.ShortenRequest
		mockSetup   func()
		expectedErr codes.Code
	}{
		{
			name: "Success",
			req: &shlink.ShortenRequest{
				UserId: "user1",
				Url:    "https://example.com",
			},
			mockSetup: func() {
				urlService.EXPECT().
					Shorten(gomock.Any(), "user1", "https://example.com").
					Return("short-url", nil)
			},
			expectedErr: codes.OK,
		},
		{
			name: "Invalid URL",
			req: &shlink.ShortenRequest{
				UserId: "user1",
				Url:    "invalid-url",
			},
			mockSetup: func() {
				urlService.EXPECT().
					Shorten(gomock.Any(), "user1", "invalid-url").
					Return("", fmt.Errorf("invalid URL"))
			},
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "Conflict",
			req: &shlink.ShortenRequest{
				UserId: "user1",
				Url:    "https://example.com",
			},
			mockSetup: func() {
				urlService.EXPECT().
					Shorten(gomock.Any(), "user1", "https://example.com").
					Return("short-url", fmt.Errorf("conflict"))
			},
			expectedErr: codes.AlreadyExists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			resp, err := client.Shorten(ctx, tc.req)

			if tc.expectedErr == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "short-url", resp.ShortUrl)
			} else {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedErr, statusErr.Code())
			}
		})
	}
}

func TestGRPCServer_Redirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis := bufconn.Listen(bufSize)
	defer lis.Close()

	urlService := url.NewMockIURLService(ctrl)
	healthService := health.NewMockIHealthService(ctrl)

	server, err := NewGRPCServer(urlService, healthService)
	require.NoError(t, err)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := shlink.NewShlinkServiceClient(conn)

	testCases := []struct {
		name        string
		req         *shlink.RedirectRequest
		mockSetup   func()
		expectedErr codes.Code
	}{
		{
			name: "Success",
			req: &shlink.RedirectRequest{
				Id: "valid-id",
			},
			mockSetup: func() {
				urlService.EXPECT().
					GetOriginal(gomock.Any(), "valid-id").
					Return("https://example.com", nil)
			},
			expectedErr: codes.OK,
		},
		{
			name: "URL is deleted",
			req: &shlink.RedirectRequest{
				Id: "deleted-id",
			},
			mockSetup: func() {
				urlService.EXPECT().
					GetOriginal(gomock.Any(), "deleted-id").
					Return("", errors.New("URL is deleted"))
			},
			expectedErr: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			resp, err := client.Redirect(ctx, tc.req)

			if tc.expectedErr == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "https://example.com", resp.OriginalUrl)
			} else {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedErr, statusErr.Code())
			}
		})
	}
}

func TestGRPCServer_ShortenJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis := bufconn.Listen(bufSize)
	defer lis.Close()

	urlService := url.NewMockIURLService(ctrl)
	healthService := health.NewMockIHealthService(ctrl)

	server, err := NewGRPCServer(urlService, healthService)
	require.NoError(t, err)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := shlink.NewShlinkServiceClient(conn)

	testCases := []struct {
		name        string
		req         *shlink.ShortenJSONRequest
		mockSetup   func()
		expectedErr codes.Code
	}{
		{
			name: "Success",
			req: &shlink.ShortenJSONRequest{
				UserId: "user1",
				Url:    "https://example.com",
			},
			mockSetup: func() {
				urlService.EXPECT().
					Shorten(gomock.Any(), "user1", "https://example.com").
					Return("short-url", nil)
			},
			expectedErr: codes.OK,
		},
		{
			name: "Conflict",
			req: &shlink.ShortenJSONRequest{
				UserId: "user1",
				Url:    "https://example.com",
			},
			mockSetup: func() {
				urlService.EXPECT().
					Shorten(gomock.Any(), "user1", "https://example.com").
					Return("short-url", fmt.Errorf("conflict"))
			},
			expectedErr: codes.AlreadyExists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			resp, err := client.ShortenJSON(ctx, tc.req)

			if tc.expectedErr == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "short-url", resp.ShortUrl)
			} else {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedErr, statusErr.Code())
			}
		})
	}
}

func TestGRPCServer_ShortenJSONBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis := bufconn.Listen(bufSize)
	defer lis.Close()

	urlService := url.NewMockIURLService(ctrl)
	healthService := health.NewMockIHealthService(ctrl)

	server, err := NewGRPCServer(urlService, healthService)
	require.NoError(t, err)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := shlink.NewShlinkServiceClient(conn)

	testCases := []struct {
		name        string
		req         *shlink.ShortenJSONBatchRequest
		mockSetup   func()
		expectedErr codes.Code
	}{
		{
			name: "Success",
			req: &shlink.ShortenJSONBatchRequest{
				UserId: "user1",
				Urls:   []string{"https://example.com"},
			},
			mockSetup: func() {
				urlService.EXPECT().
					ShortenList(gomock.Any(), "user1", gomock.Any()).
					Return(dto.BatchShortenResponseDTO{
						{CorrelationID: "1", ShortURL: "short-url"},
					}, nil)
			},
			expectedErr: codes.OK,
		},
		{
			name: "Error",
			req: &shlink.ShortenJSONBatchRequest{
				UserId: "user1",
				Urls:   []string{"https://example.com"},
			},
			mockSetup: func() {
				urlService.EXPECT().
					ShortenList(gomock.Any(), "user1", gomock.Any()).
					Return(nil, fmt.Errorf("internal error"))
			},
			expectedErr: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			resp, err := client.ShortenJSONBatch(ctx, tc.req)

			if tc.expectedErr == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "short-url", resp.ShortUrls[0])
			} else {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedErr, statusErr.Code())
			}
		})
	}
}

func TestGRPCServer_GetUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis := bufconn.Listen(bufSize)
	defer lis.Close()

	urlService := url.NewMockIURLService(ctrl)
	healthService := health.NewMockIHealthService(ctrl)

	server, err := NewGRPCServer(urlService, healthService)
	require.NoError(t, err)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := shlink.NewShlinkServiceClient(conn)

	testCases := []struct {
		name        string
		req         *shlink.GetUserURLsRequest
		mockSetup   func()
		expectedErr codes.Code
	}{
		{
			name: "Success",
			req: &shlink.GetUserURLsRequest{
				UserId: "user1",
			},
			mockSetup: func() {
				urlService.EXPECT().
					GetUserURLs(gomock.Any(), "user1").
					Return(dto.GetUserURLsResponseDTO{
						{ShortURL: "short-url", OriginalURL: "https://example.com"},
					}, nil)
			},
			expectedErr: codes.OK,
		},
		{
			name: "No URLs",
			req: &shlink.GetUserURLsRequest{
				UserId: "user1",
			},
			mockSetup: func() {
				urlService.EXPECT().
					GetUserURLs(gomock.Any(), "user1").
					Return(dto.GetUserURLsResponseDTO{}, nil)
			},
			expectedErr: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			resp, err := client.GetUserURLs(ctx, tc.req)

			if tc.expectedErr == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if len(resp.Urls) > 0 {
					assert.Equal(t, "https://example.com", resp.Urls[0])
				}
			} else {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedErr, statusErr.Code())
			}
		})
	}
}

func TestGRPCServer_DeleteUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis := bufconn.Listen(bufSize)
	defer lis.Close()

	urlService := url.NewMockIURLService(ctrl)
	healthService := health.NewMockIHealthService(ctrl)

	server, err := NewGRPCServer(urlService, healthService)
	require.NoError(t, err)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := shlink.NewShlinkServiceClient(conn)

	testCases := []struct {
		name        string
		req         *shlink.DeleteUserURLsRequest
		mockSetup   func()
		expectedErr codes.Code
	}{
		{
			name: "Success",
			req: &shlink.DeleteUserURLsRequest{
				UserId: "user1",
				Urls:   []string{"short-url"},
			},
			mockSetup: func() {
				urlService.EXPECT().
					DeleteUserURLs(gomock.Any(), "user1", []string{"short-url"}).
					Return(nil)
			},
			expectedErr: codes.OK,
		},
		{
			name: "Unauthorized",
			req: &shlink.DeleteUserURLsRequest{
				UserId: "",
				Urls:   []string{"short-url"},
			},
			mockSetup:   func() {},
			expectedErr: codes.Unauthenticated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			resp, err := client.DeleteUserURLs(ctx, tc.req)

			if tc.expectedErr == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.Success)
			} else {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedErr, statusErr.Code())
			}
		})
	}
}

func TestGRPCServer_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis := bufconn.Listen(bufSize)
	defer lis.Close()

	urlService := url.NewMockIURLService(ctrl)
	healthService := health.NewMockIHealthService(ctrl)

	server, err := NewGRPCServer(urlService, healthService)
	require.NoError(t, err)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()
	client := shlink.NewShlinkServiceClient(conn)

	testCases := []struct {
		name           string
		mockSetup      func()
		expectedStatus string
	}{
		{
			name: "Success",
			mockSetup: func() {
				healthService.EXPECT().
					CheckDatabaseConnection(gomock.Any()).
					Return(nil)
			},
			expectedStatus: "OK",
		},
		{
			name: "Database error",
			mockSetup: func() {
				healthService.EXPECT().
					CheckDatabaseConnection(gomock.Any()).
					Return(fmt.Errorf("database error"))
			},
			expectedStatus: "Database connection error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			resp, err := client.Ping(ctx, &shlink.PingRequest{})

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tc.expectedStatus, resp.Status)

		})
	}
}

func TestGRPCServer_GetStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis := bufconn.Listen(bufSize)
	defer lis.Close()

	urlService := url.NewMockIURLService(ctrl)
	healthService := health.NewMockIHealthService(ctrl)

	server, err := NewGRPCServer(urlService, healthService)
	require.NoError(t, err)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := shlink.NewShlinkServiceClient(conn)

	testCases := []struct {
		name        string
		mockSetup   func()
		metadata    map[string]string
		expectedErr codes.Code
	}{
		{
			name: "Success",
			mockSetup: func() {
				urlService.EXPECT().
					IsAllowed("192.168.1.1").
					Return(true)
				urlService.EXPECT().
					GetStats(gomock.Any()).
					Return(map[string]int{
						"urls":  10,
						"users": 5,
					}, nil)
			},
			metadata: map[string]string{
				"x-real-ip": "192.168.1.1",
			},
			expectedErr: codes.OK,
		},
		{
			name: "Metadata is missing",
			mockSetup: func() {
			},
			metadata:    nil, // Нет метаданных
			expectedErr: codes.PermissionDenied,
		},
		{
			name: "X-Real-IP header is missing",
			mockSetup: func() {
			},
			metadata: map[string]string{
				"other-header": "value",
			},
			expectedErr: codes.PermissionDenied,
		},
		{
			name: "Access forbidden",
			mockSetup: func() {
				urlService.EXPECT().
					IsAllowed("192.168.1.1").
					Return(false)
			},
			metadata: map[string]string{
				"x-real-ip": "192.168.1.1",
			},
			expectedErr: codes.PermissionDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			ctx := context.Background()
			if tc.metadata != nil {
				md := metadata.New(tc.metadata)
				ctx = metadata.NewOutgoingContext(ctx, md)
			}

			resp, err := client.GetStats(ctx, &shlink.GetStatsRequest{})

			if tc.expectedErr == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, int64(10), resp.TotalUrls)
				assert.Equal(t, int64(5), resp.TotalUsers)
			} else {
				assert.Error(t, err)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedErr, statusErr.Code())
			}
		})
	}
}
