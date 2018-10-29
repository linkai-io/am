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

func NewLocalStorage(prefixPath string) *LocalStorage {
	return &LocalStorage{prefixPath: prefixPath}
}

func (s *LocalStorage) Init(config []byte) error {
	return os.MkdirAll(s.prefixPath, 0766)
}

// Writes the data to local storage, returning the hash and link/path
func (s *LocalStorage) Write(ctx context.Context, address *am.ScanGroupAddress, data []byte) (string, string, error) {
	if data == nil || len(data) == 0 {
		return "", "", nil
	}
	
	hashName := convert.HashData(data)
	fileName := PathFromData(address, hashName)
	if fileName == "null" {
		return "", "", nil
	}
	dir := filepath.Dir(s.prefixPath + fileName)
	if err := os.MkdirAll(dir, 0766); err != nil {
		return "", "", err
	}
	err := ioutil.WriteFile(s.prefixPath+fileName, data, 0766)
	return hashName, s.prefixPath + fileName, err
}
