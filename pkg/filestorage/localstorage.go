package filestorage

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
)

type LocalStorage struct {
	prefixPath string
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{}
}

func (s *LocalStorage) Init() error {
	return nil
}

// Writes the data to local storage, returning the hash and link/path
func (s *LocalStorage) Write(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, data []byte) (string, string, error) {
	if data == nil || len(data) == 0 {
		return "", "", nil
	}

	hashName := convert.HashData(data)
	fileName := PathFromData(address, hashName)
	if fileName == "null" {
		return "", "", nil
	}
	dir := filepath.Dir(userContext.GetOrgCID() + fileName)
	if err := os.MkdirAll(dir, 0766); err != nil {
		return "", "", err
	}
	err := ioutil.WriteFile(userContext.GetOrgCID()+fileName, data, 0766)
	return hashName, userContext.GetOrgCID() + fileName, err
}

func (s *LocalStorage) GetInfraFile(ctx context.Context, pathName, objectName string) ([]byte, error) {
	return nil, nil
}

func (s *LocalStorage) PutInfraFile(ctx context.Context, pathName, objectName string, data []byte) error {
	return nil
}
