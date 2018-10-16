package filestorage

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/linkai-io/am/am"
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

func (s *LocalStorage) Write(ctx context.Context, address *am.ScanGroupAddress, data []byte) error {
	fileName := PathFromData(address, data)
	if fileName == "null" {
		return nil
	}
	dir := filepath.Dir(s.prefixPath + fileName)
	if err := os.MkdirAll(dir, 0766); err != nil {
		return err
	}
	return ioutil.WriteFile(s.prefixPath+fileName, data, 0766)
}
