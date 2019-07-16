package secrets

type Secrets interface {
	WithCredentials(id, key string)
	GetSecureParameter(key string) ([]byte, error)
	SetSecureParameter(key, value string) error
}
