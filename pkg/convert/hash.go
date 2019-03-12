package convert

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

// HashAddress for ip and host returning a hash key to allow modules to check if hosts exist
func HashAddress(ipAddress, host string) string {
	hash := md5.New()
	hash.Write([]byte(strings.TrimSpace(strings.ToLower(ipAddress))))
	hash.Write([]byte(strings.TrimSpace(strings.ToLower(host))))
	sum := hash.Sum(nil)
	return hex.EncodeToString(sum)
}

// HashData using a sha1 hash
func HashData(data []byte) string {
	hash := sha1.New()
	hash.Write(data)
	result := hash.Sum(nil)
	return hex.EncodeToString(result)
}
