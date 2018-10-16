package convert

import (
	"crypto/sha1"
	"encoding/hex"
	"strconv"

	"github.com/spaolacci/murmur3"
)

// HashAddress for ip and host returning a hash key to allow modules to check if hosts exist
func HashAddress(ipAddress, host string) string {
	hash := murmur3.New64()
	hash.Write([]byte(ipAddress))
	hash.Write([]byte(host))
	sum := hash.Sum64()
	return strconv.FormatUint(sum, 10)
}

// HashData using a sha1 hash
func HashData(data []byte) string {
	hash := sha1.New()
	hash.Write(data)
	result := hash.Sum(nil)
	return hex.EncodeToString(result)
}
