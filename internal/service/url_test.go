package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	repo "github.com/GlebRadaev/shlink/internal/repository"
	repository "github.com/GlebRadaev/shlink/internal/repository/inmemory"
	"github.com/stretchr/testify/assert"
)

var globalCfg *config.Config
var globalErr error

func setup() (repo.Repository, *URLService, *config.Config, error) {
	memStorage := repository.NewMemoryStorage()
	if globalCfg == nil && globalErr == nil {
		globalCfg, globalErr = config.ParseAndLoadConfig()
	}
	if globalErr != nil {
		return nil, nil, nil, globalErr
	}
	urlService := NewURLService(memStorage, globalCfg)
	return memStorage, urlService, globalCfg, nil
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
			args:    args{url: string(make([]byte, MaxURLLength+1))},
			wantErr: errors.New("invalid URL format"),
		},
		// {
		// 	name:    "invalid domain name",
		// 	args:    args{url: "http://invalid_domain"},
		// 	wantErr: errors.New("invalid domain name"),
		// },
	}

	memStorage, urlService, cfg, err := setup()
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
				assert.Equal(t, len(cfg.BaseURL)+1+MaxIDLength, len(got), "Expected ID length to be %d, but got %d", len(cfg.BaseURL)+1+MaxIDLength, len(got))
				storedURL, found := memStorage.Get(strings.Split(got, "/")[len(strings.Split(got, "/"))-1])
				assert.True(t, found, "Expected the ID to be stored, but it was not found")
				assert.Equal(t, tt.args.url, storedURL, "Stored URL mismatch: got %v, want %v", storedURL, tt.args.url)
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
		setup   func(storage *repository.MemoryStorage)
		want    string
		wantErr error
	}{
		{
			name: "valid ID",
			args: args{id: "testIDid"},
			setup: func(storage *repository.MemoryStorage) {
				_ = storage.AddURL("testIDid", "http://example.com")
			},
			want:    "http://example.com",
			wantErr: nil,
		},
		{
			name:    "invalid ID length",
			args:    args{id: "short"},
			setup:   func(storage *repository.MemoryStorage) {},
			want:    "",
			wantErr: errors.New("invalid ID"),
		},
		{
			name:    "invalid ID format",
			args:    args{id: "invalid!"},
			setup:   func(storage *repository.MemoryStorage) {},
			want:    "",
			wantErr: errors.New("invalid ID"),
		},
		{
			name:    "ID not found",
			args:    args{id: "notfound"},
			setup:   func(storage *repository.MemoryStorage) {},
			want:    "",
			wantErr: errors.New("URL not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewMemoryStorage().(*repository.MemoryStorage)
			s := &URLService{
				storage: storage,
			}
			if tt.setup != nil {
				tt.setup(storage)
			}
			got, err := s.GetOriginal(tt.args.id)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
