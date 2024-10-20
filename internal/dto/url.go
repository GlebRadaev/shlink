package dto

type ShortenJSONRequestDTO struct {
	URL string `json:"url"`
}

type ShortenJSONResponseDTO struct {
	Result string `json:"result"`
}
