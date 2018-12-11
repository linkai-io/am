package secrets

import (
	"fmt"
	"strconv"
	"strings"
)

// SecretsCache for accessing cached/stored secrets
type SecretsCache struct {
	Region      string
	Environment string
	secrets     Secrets
}

// NewSecretsCache returns an instance for acquiring the secrets from either local env vars or AWS
func NewSecretsCache(env, region string) *SecretsCache {
	s := &SecretsCache{Environment: env, Region: region}
	if s.Environment != "local" {
		s.secrets = NewAWSSecrets(region)
	} else {
		s.secrets = NewEnvSecrets()
	}
	return s
}

// GetSecureString allows caller to provide the full key to return a string value
func (s *SecretsCache) GetSecureString(key string) (string, error) {
	data, err := s.secrets.GetSecureParameter(key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DBString returns the database connection string for the environment and service
func (s *SecretsCache) DBString(serviceKey string) (string, error) {
	data, err := s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/db/%s/dbstring", s.Environment, serviceKey))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ServicePassword retrieves the password for the initialized servicekey
func (s *SecretsCache) ServicePassword(serviceKey string) (string, error) {
	data, err := s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/db/%s/pwd", s.Environment, serviceKey))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *SecretsCache) StateAddr() (string, error) {
	data, err := s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/state/config", s.Environment))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *SecretsCache) StatePassword() (string, error) {
	data, err := s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/state/pwd", s.Environment))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *SecretsCache) DiscoveryAddr() (string, error) {
	data, err := s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/discovery/config", s.Environment))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *SecretsCache) LoadBalancerAddr() (string, error) {
	data, err := s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/loadbalancer/config", s.Environment))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *SecretsCache) WebFilePath() (string, error) {
	data, err := s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/webfilepath", s.Environment))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *SecretsCache) DNSAddresses() ([]string, error) {
	data, err := s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/dnsaddresses", s.Environment))
	if err != nil {
		return []string{""}, err
	}
	hosts := strings.Trim(string(data), " ")
	return strings.Split(hosts, ","), nil
}

func (s *SecretsCache) SystemOrgID() (int, error) {
	data, err := s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/system/orgid", s.Environment))
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(string(data))
}

func (s *SecretsCache) SystemUserID() (int, error) {
	data, err := s.secrets.GetSecureParameter(fmt.Sprintf("/am/%s/system/userid", s.Environment))
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(string(data))
}

func (s *SecretsCache) SetSystemIDs(orgID, userID int) error {
	if err := s.secrets.SetSecureParameter(fmt.Sprintf("/am/%s/system/orgid", s.Environment), strconv.Itoa(orgID)); err != nil {
		return err
	}
	return s.secrets.SetSecureParameter(fmt.Sprintf("/am/%s/system/userid", s.Environment), strconv.Itoa(userID))
}
