package dto

// ShortenJSONRequestDTO defines the structure of the shorten request
type ShortenJSONRequestDTO struct {
	URL string `json:"url"`
}

// ShortenJSONResponseDTO defines the structure of the shorten response
type ShortenJSONResponseDTO struct {
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

// GetUserURLsResponse defines the structure of the get user URLs response
type GetUserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type GetUserURLsResponseDTO []GetUserURLsResponse

type DeleteURLRequestDTO []string
