package filestorage

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/linkai-io/am/am"
)

var (
	ErrNameTooSmall = errors.New("name length must be greater than 5")
)

type Storage interface {
	Init(config []byte) error
	Write(ctx context.Context, address *am.ScanGroupAddress, data []byte) (string, string, error)
}

// ShardName takes an input name and shards it 5 levels
func ShardName(name string) (string, error) {
	if len(name) < 5 {
		return "", ErrNameTooSmall
	}

	var buf bytes.Buffer
	buf.Grow(len(name) + 10)

	for i := 0; i < 5; i++ {
		buf.WriteByte(byte('/'))
		buf.WriteByte(byte(name[i]))
	}
	buf.WriteByte(byte('/'))
	buf.Write([]byte(name))
	return buf.String(), nil
}

// PathFromData takes in an address and raw bytes to create a hashed / sharded
// path name
func PathFromData(address *am.ScanGroupAddress, name string) string {
	if len(name) == 0 {
		return "null"
	}

	log.Printf("name: %s\n", name)
	sharded, err := ShardName(name)
	if err != nil {
		return "null"
	}
	var buf bytes.Buffer
	buf.Write([]byte("/"))
	buf.Write([]byte(strconv.Itoa(address.OrgID)))
	buf.Write([]byte("/"))
	buf.Write([]byte(strconv.Itoa(address.GroupID)))
	buf.Write([]byte(sharded))
	return buf.String()
}
