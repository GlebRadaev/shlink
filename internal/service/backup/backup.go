// Package backup provides functionality for managing backups of URL data.
// It defines an interface `IBackupService` and implements a service `BackupService`
// that can load and save data to a backup file.
package backup

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/utils"
)

// IBackupService defines the methods for loading and saving backup data.
type IBackupService interface {
	// LoadData loads backup data from a file.
	// It returns a map where the key is the short URL and the value is the original URL.
	LoadData() (map[string]string, error)

	// SaveData saves a map of short URLs and original URLs to a backup file.
	// It will overwrite existing data in the backup file.
	SaveData(data map[string]string) error
}

// BackupService provides the implementation of IBackupService.
type BackupService struct {
	// filename is the path to the backup file.
	filename string
}

// NewBackupService creates a new instance of BackupService.
func NewBackupService(filename string) *BackupService {
	return &BackupService{filename: filename}
}

// LoadData loads backup data from the backup file.
func (b *BackupService) LoadData() (map[string]string, error) {
	file, err := os.Open(b.filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]string), nil
		}
		return nil, err
	}
	defer file.Close()

	data := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var url dto.URLFileDataDTO
		if err := json.Unmarshal(scanner.Bytes(), &url); err != nil {
			return nil, err
		}
		data[url.ShortURL] = url.OriginalURL
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

// SaveData saves a map of short URLs to original URLs in the backup file.
func (b *BackupService) SaveData(data map[string]string) error {
	existingData, err := b.LoadData()
	if err != nil {
		return err
	}

	for shortURL, originalURL := range data {
		existingData[shortURL] = originalURL
	}

	file, err := os.OpenFile(b.filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for shortURL, originalURL := range data {
		urlDTO := dto.URLFileDataDTO{UUID: utils.GenerateUUID(), ShortURL: shortURL, OriginalURL: originalURL}
		jsonData, err := json.Marshal(urlDTO)
		if err != nil {
			return err
		}
		if _, err := writer.WriteString(string(jsonData) + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}
