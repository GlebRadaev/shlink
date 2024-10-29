package url_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service"
	"github.com/GlebRadaev/shlink/internal/service/url"

	"github.com/stretchr/testify/assert"
)

func setup() (interfaces.Repository, *url.URLService, *config.Config, error) {
	cfg, err := config.ParseAndLoadConfig()
	if err != nil {
		return nil, nil, nil, err
	}
	log, _ := logger.NewLogger("info")
	repositories := repository.NewRepositoryFactory(cfg)
	services := service.NewServiceFactory(cfg, log, repositories)
	return repositories.MemoryRepo, services.URLService, cfg, nil
}

func TestURLService_Shorten(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "valid http URL",
			args:    args{url: "http://example.com"},
			wantErr: nil,
		},
		{
			name:    "valid https URL",
			args:    args{url: "https://example.com"},
			wantErr: nil,
		},
		{
			name:    "invalid URL scheme",
			args:    args{url: "httpsss://example.com"},
			wantErr: errors.New("invalid URL scheme"),
		},
		{
			name:    "invalid URL format (URL too long)",
			args:    args{url: string(make([]byte, 2048+1))},
			wantErr: errors.New("invalid URL format"),
		},
	}

	memoryRepo, urlService, cfg, err := setup()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := urlService.Shorten(tt.args.url)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(cfg.BaseURL)+1+8, len(got), "Expected ID length to be %d, but got %d", len(cfg.BaseURL)+1+8, len(got))

				storedURL, found := memoryRepo.Get(strings.Split(got, "/")[len(strings.Split(got, "/"))-1])
				assert.True(t, found, "Expected the ID to be stored in memoryRepo, but it was not found")
				assert.Equal(t, tt.args.url, storedURL, "Stored URL mismatch in memoryRepo: got %v, want %v", storedURL, tt.args.url)
			}
		})
	}
}

func TestURLService_GetOriginal(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		setup   map[string]string
		want    string
		wantErr error
	}{
		{
			name: "valid ID",
			args: args{id: "testIDid"},
			setup: map[string]string{
				"testIDid": "http://example.com",
			},
			want:    "http://example.com",
			wantErr: nil,
		},
		{
			name:    "invalid ID length",
			args:    args{id: "short"},
			setup:   map[string]string{},
			want:    "",
			wantErr: errors.New("invalid ID"),
		},
		{
			name:    "invalid ID format",
			args:    args{id: "invalid!"},
			setup:   map[string]string{},
			want:    "",
			wantErr: errors.New("invalid ID"),
		},
		{
			name:    "ID not found",
			args:    args{id: "notfound"},
			setup:   map[string]string{},
			want:    "",
			wantErr: errors.New("URL not found"),
		},
	}

	memStorage, urlService, _, err := setup()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, url := range tt.setup {
				_ = memStorage.AddURL(key, url)
			}
			got, err := urlService.GetOriginal(tt.args.id)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
