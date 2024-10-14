package repository

type Repository interface {
	AddURL(string, string) error
	Get(string) (string, bool)
}
