package autocertcache

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/acme/autocert"
)

// GroupDirCache implements Cache using a directory on the local filesystem.
// If the directory does not exist, it will be created with 0770 permissions.
type GroupDirCache string

// Get reads a certificate data from the specified file name.
func (d GroupDirCache) Get(ctx context.Context, name string) ([]byte, error) {
	name = filepath.Join(string(d), name)
	var (
		data []byte
		err  error
		done = make(chan struct{})
	)
	go func() {
		data, err = ioutil.ReadFile(name)
		close(done)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
	}
	if os.IsNotExist(err) {
		return nil, autocert.ErrCacheMiss
	}
	return data, err
}

// Put writes the certificate data to the specified file name.
// The file will be created with 0660 permissions.
func (d GroupDirCache) Put(ctx context.Context, name string, data []byte) error {
	if err := os.MkdirAll(string(d), 0770); err != nil {
		return err
	}

	done := make(chan struct{})
	var err error
	go func() {
		defer close(done)
		var tmp string
		if tmp, err = d.writeTempFile(name, data); err != nil {
			return
		}
		select {
		case <-ctx.Done():
			// Don't overwrite the file if the context was canceled.
		default:
			newName := filepath.Join(string(d), name)
			err = os.Rename(tmp, newName)
			if err == nil {
				err = os.Chmod(newName, 0660)
			}
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	return err
}

// Delete removes the specified file name.
func (d GroupDirCache) Delete(ctx context.Context, name string) error {
	name = filepath.Join(string(d), name)
	var (
		err  error
		done = make(chan struct{})
	)
	go func() {
		err = os.Remove(name)
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// writeTempFile writes b to a temporary file, closes the file and returns its path.
func (d GroupDirCache) writeTempFile(prefix string, b []byte) (string, error) {
	// TempFile uses 0600 permissions will Chmod after creation
	f, err := ioutil.TempFile(string(d), prefix)
	if err != nil {
		return "", err
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return "", err
	}
	return f.Name(), f.Close()
}
