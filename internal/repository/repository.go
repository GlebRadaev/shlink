package repository

type Repository interface {
	Add(string, string) error
	Get(string) (string, bool)
}
