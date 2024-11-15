package backup_test

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/service/backup"
	"github.com/stretchr/testify/assert"
)

func setupBackupServiceTest(t *testing.T, content string) (*backup.BackupService, string) {
	tempDir := t.TempDir()
	file, err := os.CreateTemp(tempDir, "backup_test.txt")
	if err != nil {
		t.Fatalf("Failed to create a temporary file: %v", err)
	}
	if content != "" {
		if _, err := file.WriteString(content); err != nil {
			t.Fatalf("Failed to write to the temporary file: %v", err)
		}
	}
	file.Close()
	service := backup.NewBackupService(file.Name())
	return service, file.Name()
}

func TestBackupService_LoadData(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   string
		setupFunc     func(filename string) // Функция для настройки условий теста
		expectedData  map[string]string
		expectedError error
	}{
		{
			name:        "Load valid data",
			fileContent: `{"uuid":"1730235510864975546","short_url":"vgWhy9ow","original_url":"http://example.com"}` + "\n",
			expectedData: map[string]string{
				"vgWhy9ow": "http://example.com",
			},
			expectedError: nil,
		},
		{
			name:          "File not found",
			setupFunc:     func(filename string) {}, // Не создаем файл для теста отсутствующего файла
			expectedData:  map[string]string{},
			expectedError: nil,
		},
		{
			name:          "Invalid JSON in file",
			fileContent:   `invalid json data`,
			expectedData:  nil,
			expectedError: errors.New("invalid character 'i' looking for beginning of value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var service *backup.BackupService

			// Настройка тестовой среды
			if tt.setupFunc != nil {
				tempDir := t.TempDir()
				filename := tempDir + "/backup_test.txt"
				service = backup.NewBackupService(filename)
				tt.setupFunc(filename)
			} else {
				// В остальных случаях создаем временный файл
				service, _ = setupBackupServiceTest(t, tt.fileContent)
			}

			data, err := service.LoadData()
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, data)
			}
		})
	}
}

func TestBackupService_SaveData(t *testing.T) {
	tests := []struct {
		name          string
		inputData     map[string]string
		expectedError error
	}{
		{
			name: "Successful save",
			inputData: map[string]string{
				"short1": "https://example.com",
			},
			expectedError: nil,
		},
		{
			name: "Failed save (invalid path)",
			inputData: map[string]string{
				"short1": "https://example.com",
			},
			expectedError: errors.New("open /invalid_path/file.txt: no such file or directory"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var service *backup.BackupService
			var filename string
			if tt.expectedError == nil {
				service, filename = setupBackupServiceTest(t, "")
			} else {
				filename = "/invalid_path/file.txt"
				service = backup.NewBackupService(filename)
			}
			err := service.SaveData(tt.inputData)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				content, _ := os.ReadFile(filename)
				var urlDTO dto.URLFileDataDTO
				err = json.Unmarshal(content[:len(content)-1], &urlDTO)
				assert.NoError(t, err)
				assert.Equal(t, "short1", urlDTO.ShortURL)
				assert.Equal(t, "https://example.com", urlDTO.OriginalURL)
			}
		})
	}
}
