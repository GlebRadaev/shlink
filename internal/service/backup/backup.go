package backup

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/utils"
)

type BackupService struct {
	filename string
}

func NewBackupService(filename string) *BackupService {
	return &BackupService{filename: filename}
}

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
		var url dto.URLDTO
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

func (b *BackupService) SaveData(data map[string]string) error {
	file, err := os.OpenFile(b.filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for shortURL, originalURL := range data {
		urlDTO := dto.URLDTO{UUID: utils.GenerateUUID(), ShortURL: shortURL, OriginalURL: originalURL}
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
