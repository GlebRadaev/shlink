package repository

type Repository interface {
	AddUrl(string, string) error
	Get(string) (string, bool)
}
