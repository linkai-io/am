package secrets

import (
	"fmt"
)

// DBSecrets for accessing database connection strings
type DBSecrets struct {
	Region      string
	ServiceKey  string
	Environment string
	secrets     Secrets
}

// NewDBSecrets returns an instance for acquiring the database connection string from
// either local env vars or AWS
func NewDBSecrets(env, serviceKey, region string) *DBSecrets {
	s := &DBSecrets{Environment: env, ServiceKey: serviceKey, Region: region}
	if s.Environment != "local" {
		s.secrets = NewAWSSecrets(region)
	} else {
		s.secrets = NewEnvSecrets()
	}
	return s
}

// DBString returns the database connection string for the environment and service
func (s *DBSecrets) DBString() ([]byte, error) {
	return s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/db/%s/dbstring", s.Environment, s.ServiceKey))
}
