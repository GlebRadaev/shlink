package dto

// ShortenJSONRequestDTO defines the structure of a single shorten URL request payload.
type ShortenJSONRequestDTO struct {
	URL string `json:"url"` // The original URL to be shortened.
}

// ShortenJSONResponseDTO defines the structure of the response for a single shorten URL request.
type ShortenJSONResponseDTO struct {
	Result string `json:"result"` // The shortened URL.
}

// URLFileDataDTO defines the structure of URL data stored in the file.
type URLFileDataDTO struct {
	UUID        string `json:"uuid"`         // Unique identifier for the URL.
	ShortURL    string `json:"short_url"`    // The shortened URL.
	OriginalURL string `json:"original_url"` // The original URL.
}

// BatchShortenRequest represents a single URL shorten request in a batch operation.
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"` // Identifier to correlate the request with the response.
	OriginalURL   string `json:"original_url"`   // The original URL to be shortened.
}

// BatchShortenRequestDTO represents a list of batch shorten requests.
type BatchShortenRequestDTO []BatchShortenRequest

// BatchShortenResponse represents a single URL shorten response in a batch operation.
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"` // Identifier to correlate the response with the request.
	ShortURL      string `json:"short_url"`      // The shortened URL.
}

// BatchShortenResponseDTO represents a list of batch shorten responses.
type BatchShortenResponseDTO []BatchShortenResponse

// GetUserURLsResponse defines the structure of a single user's shortened URL entry.
type GetUserURLsResponse struct {
	ShortURL    string `json:"short_url"`    // The shortened URL.
	OriginalURL string `json:"original_url"` // The original URL.
}

// GetUserURLsResponseDTO represents a list of user's shortened URL entries.
type GetUserURLsResponseDTO []GetUserURLsResponse

// DeleteURLRequestDTO represents a list of shortened URL IDs to be deleted.
type DeleteURLRequestDTO []string
