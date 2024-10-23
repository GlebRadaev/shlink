package dto

type ShortenJSONRequestDTO struct {
	URL string `json:"url"`
}

type ShortenJSONResponseDTO struct {
	Result string `json:"result"`
}

type URLDTO struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
