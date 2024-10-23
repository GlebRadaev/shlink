package interfaces

type Repository interface {
	AddURL(string, string) error
	Get(string) (string, bool)
	GetAll() map[string]string
}
