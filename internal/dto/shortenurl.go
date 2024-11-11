package dto

// ShortenRequestDTO defines the structure of the shorten request
type ShortenRequestDTO struct {
	URL string `json:"url"`
}

// ShortenResponseDTO defines the structure of the shorten response
type ShortenResponseDTO struct {
	Result string `json:"result"`
}

// URLFileDataDTO defines the structure of the data in the file
type URLFileDataDTO struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// BatchShortenRequest defines the structure of the batch shorten request
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShortenRequestDTO []BatchShortenRequest

// BatchShortenResponse defines the structure of the batch shorten response
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type BatchShortenResponseDTO []BatchShortenResponse
