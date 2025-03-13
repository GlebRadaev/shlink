package api

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/service/health"
	"github.com/GlebRadaev/shlink/internal/service/url"
	"github.com/GlebRadaev/shlink/pkg/shlink"
)

// grpcServer implements the methods of the gRPC server.
type grpcServer struct {
	shlink.UnimplementedShlinkServiceServer
	urlService    url.IURLService
	healthService health.IHealthService
}

// NewGRPCServer creates and returns a new instance of the gRPC server.
func NewGRPCServer(urlService url.IURLService, healthService health.IHealthService) (*grpc.Server, error) {
	server := grpc.NewServer()
	shlink.RegisterShlinkServiceServer(server, &grpcServer{
		urlService:    urlService,
		healthService: healthService,
	})
	reflection.Register(server)

	return server, nil
}

// Shorten handles the gRPC request to shorten a URL.
func (s *grpcServer) Shorten(ctx context.Context, req *shlink.ShortenRequest) (*shlink.ShortenResponse, error) {
	shortURL, err := s.urlService.Shorten(ctx, req.UserId, req.Url)
	if err != nil {
		if strings.Contains(err.Error(), "invalid URL") {
			return nil, status.Error(codes.InvalidArgument, "invalid URL")
		}
		if strings.Contains(err.Error(), "conflict") {
			return &shlink.ShortenResponse{ShortUrl: shortURL}, status.Error(codes.AlreadyExists, "URL already shortened")
		}
		return nil, err
	}
	return &shlink.ShortenResponse{ShortUrl: shortURL}, nil
}

// Redirect handles the gRPC request to redirect to the original URL.
func (s *grpcServer) Redirect(ctx context.Context, req *shlink.RedirectRequest) (*shlink.RedirectResponse, error) {
	originalURL, err := s.urlService.GetOriginal(ctx, req.Id)
	if err != nil {
		if err.Error() == "URL is deleted" {
			return nil, status.Error(codes.NotFound, "URL is deleted")
		}
		return nil, err
	}
	return &shlink.RedirectResponse{OriginalUrl: originalURL}, nil
}

// ShortenJSON handles the gRPC request to shorten a URL in JSON format.
func (s *grpcServer) ShortenJSON(ctx context.Context, req *shlink.ShortenJSONRequest) (*shlink.ShortenJSONResponse, error) {
	shortURL, err := s.urlService.Shorten(ctx, req.UserId, req.Url)
	if err != nil {
		if strings.Contains(err.Error(), "conflict") {
			return &shlink.ShortenJSONResponse{ShortUrl: shortURL}, status.Error(codes.AlreadyExists, "URL already shortened")
		}
		return nil, err
	}
	return &shlink.ShortenJSONResponse{ShortUrl: shortURL}, nil
}

// ShortenJSONBatch handles the gRPC request to shorten multiple URLs in batch.
func (s *grpcServer) ShortenJSONBatch(ctx context.Context, req *shlink.ShortenJSONBatchRequest) (*shlink.ShortenJSONBatchResponse, error) {
	batchReq := make(dto.BatchShortenRequestDTO, len(req.Urls))
	for i, url := range req.Urls {
		batchReq[i] = dto.BatchShortenRequest{OriginalURL: url}
	}

	shortenResults, err := s.urlService.ShortenList(ctx, req.UserId, batchReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	shortUrls := make([]string, len(shortenResults))
	for i, result := range shortenResults {
		shortUrls[i] = result.ShortURL
	}

	return &shlink.ShortenJSONBatchResponse{ShortUrls: shortUrls}, nil
}

// GetUserURLs handles the gRPC request to retrieve a user's list of URLs.
func (s *grpcServer) GetUserURLs(ctx context.Context, req *shlink.GetUserURLsRequest) (*shlink.GetUserURLsResponse, error) {
	urls, err := s.urlService.GetUserURLs(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return &shlink.GetUserURLsResponse{Urls: []string{}}, nil
	}

	userURLs := make([]string, len(urls))
	for i, url := range urls {
		userURLs[i] = url.OriginalURL
	}

	return &shlink.GetUserURLsResponse{Urls: userURLs}, nil
}

// DeleteUserURLs handles the gRPC request to delete a user's URLs.
func (s *grpcServer) DeleteUserURLs(ctx context.Context, req *shlink.DeleteUserURLsRequest) (*shlink.DeleteUserURLsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}
	err := s.urlService.DeleteUserURLs(ctx, req.UserId, req.Urls)
	if err != nil {
		return &shlink.DeleteUserURLsResponse{Success: false}, err
	}
	return &shlink.DeleteUserURLsResponse{Success: true}, nil
}

// Ping handles the gRPC request to check the service's health.
func (s *grpcServer) Ping(ctx context.Context, req *shlink.PingRequest) (*shlink.PingResponse, error) {
	if err := s.healthService.CheckDatabaseConnection(ctx); err != nil {
		return &shlink.PingResponse{Status: "Database connection error"}, nil
	}
	return &shlink.PingResponse{Status: "OK"}, nil
}

// GetStats handles the gRPC request to retrieve service statistics.
func (s *grpcServer) GetStats(ctx context.Context, req *shlink.GetStatsRequest) (*shlink.GetStatsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Metadata is missing")
	}

	ipHeaders := md.Get("x-real-ip")
	if len(ipHeaders) == 0 {
		return nil, status.Error(codes.PermissionDenied, "X-Real-IP header is missing")
	}
	clientIP := ipHeaders[0]

	if !s.urlService.IsAllowed(clientIP) {
		return nil, status.Error(codes.PermissionDenied, "Access forbidden")
	}

	stats, err := s.urlService.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	return &shlink.GetStatsResponse{
		TotalUrls:  int64(stats["urls"]),
		TotalUsers: int64(stats["users"]),
	}, nil
}
