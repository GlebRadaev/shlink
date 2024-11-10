package dto

type ShortenRequestDTO struct {
	URL string `json:"url"`
}

type ShortenResponseDTO struct {
	Result string `json:"result"`
}

type URLFileDataDTO struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
