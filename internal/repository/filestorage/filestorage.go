package filestorage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/utils"
)

type FileStorage struct {
	filename string
}

func NewFileStorage(filename string) interfaces.Repository {
	return &FileStorage{filename: filename}
}

func (s *FileStorage) AddURL(key, value string) error {
	if key == "" {
		return errors.New("shortID cannot be empty")
	}
	file, err := os.OpenFile(s.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.New("failed to open file for writing")
	}
	defer file.Close()
	newData := dto.URLDTO{
		UUID:        utils.GenerateUUID(),
		ShortURL:    key,
		OriginalURL: value,
	}
	jsonData, err := json.Marshal(newData)
	if err != nil {
		return errors.New("failed to encode data to JSON")
	}
	_, err = file.WriteString(string(jsonData) + "\n")
	return err
}

func (s *FileStorage) Get(id string) (string, bool) {
	data, err := s.loadAll()
	if err != nil {
		return "", false
	}
	var url string
	var found bool
	for _, entry := range data {
		if entry.ShortURL == id {
			url = entry.OriginalURL
			found = true
			break
		}
	}
	return url, found
}

func (s *FileStorage) GetAll() map[string]string {
	result := make(map[string]string)
	data, err := s.loadAll()
	if err != nil {
		return map[string]string{}
	}
	for _, entry := range data {
		result[entry.ShortURL] = entry.OriginalURL
	}
	return result
}

func (s *FileStorage) loadAll() ([]dto.URLDTO, error) {
	file, err := os.Open(s.filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []dto.URLDTO{}, nil
		}
		return nil, err
	}
	defer file.Close()
	var urlList []dto.URLDTO
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var url dto.URLDTO
		if err := json.Unmarshal(scanner.Bytes(), &url); err != nil {
			return nil, err
		}
		urlList = append(urlList, url)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return urlList, nil
}
