package secrets

type Secrets interface {
	GetSecureParameter(key string) ([]byte, error)
}
