package dto

import (
	"encoding/json"
	"errors"
	"io"
)

type ShortenJSONRequestDTO struct {
	URL string `json:"url"`
}

type ShortenJSONResponseDTO struct {
	Result string `json:"result"`
}

func (dto *ShortenJSONRequestDTO) ValidateRequest(r io.Reader) error {
	if r == nil {
		return errors.New("empty request body")
	}
	if err := json.NewDecoder(r).Decode(dto); err != nil {
		return errors.New("Cannot decode request")
	}
	if dto.URL == "" {
		return errors.New("URL is required")
	}
	return nil
}

type URLDTO struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
