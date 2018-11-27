package secrets

type Secrets interface {
	GetSecureParameter(key string) ([]byte, error)
	SetSecureParameter(key, value string) error
}
