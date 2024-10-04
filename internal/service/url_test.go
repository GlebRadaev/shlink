package service

import (
	"errors"
	"log"
	"testing"

	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/stretchr/testify/assert"
)

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
			name:    "invalid URL (missing scheme)",
			args:    args{url: "example.com"},
			wantErr: errors.New("invalid URL format"),
		},
		{
			name:    "too long URL",
			args:    args{url: string(make([]byte, MaxURLLength+1))},
			wantErr: errors.New("URL is too long"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewMemoryStorage()
			s := &URLService{
				storage: storage,
			}
			got, err := s.Shorten(tt.args.url)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, MaxIdLength, len(got), "Expected ID length to be %d, but got %d", MaxIdLength, len(got))
				storedURL, found := storage.Find(got)
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
				_ = storage.Save("testIDid", "http://example.com")
			},
			want:    "http://example.com",
			wantErr: nil,
		},
		{
			name:    "invalid ID length",
			args:    args{id: "short"},
			setup:   func(storage *repository.MemoryStorage) {},
			want:    "",
			wantErr: errors.New("invalid ID length"),
		},
		{
			name:    "invalid ID format",
			args:    args{id: "invalid!"},
			setup:   func(storage *repository.MemoryStorage) {},
			want:    "",
			wantErr: errors.New("invalid ID format"),
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
			storage := repository.NewMemoryStorage()
			s := &URLService{
				storage: storage,
			}
			if tt.setup != nil {
				tt.setup(storage)
			}
			got, err := s.GetOriginal(tt.args.id)
			log.Print(got)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
