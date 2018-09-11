package convert

import "github.com/spaolacci/murmur3"

// HashAddress for ip and host returning a hash key to allow modules to check if hosts exist
func HashAddress(ipAddress, host string) uint64 {
	hash := murmur3.New64()
	hash.Write([]byte(ipAddress))
	hash.Write([]byte(host))
	return hash.Sum64()
}
