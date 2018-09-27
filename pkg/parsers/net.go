package parsers

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/net/publicsuffix"
)

const (
	ipv4arpafmt = "%d.%d.%d.%d.in-addr.arpa"
	ipv6arpafmt = "%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.ip6.arpa"
)

// ParseArpa parses an in-addr.arpa or ip6.arpa name to IP address.
func ParseArpa(arpa string) (string, bool) {
	arpa = FQDNTrim(arpa)
	// IPv4
	if strings.LastIndex(arpa, "in-addr.arpa") != -1 {
		return ParseIPv4Arpa(arpa)
	} else if strings.LastIndex(arpa, "ip6.arpa") != -1 {
		return ParseIPv6Arpa(arpa)
	}
	return "", false
}

// ParseIPv4Arpa uses sscanf to ensure we only get integer values for the in-addr.arpa string.
func ParseIPv4Arpa(ipv4arpa string) (string, bool) {
	bytes := make([]int, 4)
	n, err := fmt.Sscanf(ipv4arpa, ipv4arpafmt, &bytes[3], &bytes[2], &bytes[1], &bytes[0])
	if err != nil || n != 4 {
		return "", false
	}
	return fmt.Sprintf("%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3]), true
}

// ParseIPv6Arpa uses sscanf to ensure we only get integer values for the in-addr.arpa string.
func ParseIPv6Arpa(ipv4arpa string) (string, bool) {
	bytes := make([]byte, 32)
	n, err := fmt.Sscanf(ipv4arpa, ipv6arpafmt, &bytes[31], &bytes[30], &bytes[29], &bytes[28], &bytes[27], &bytes[26], &bytes[25],
		&bytes[24], &bytes[23], &bytes[22], &bytes[21], &bytes[20], &bytes[19], &bytes[18], &bytes[17], &bytes[16], &bytes[15],
		&bytes[14], &bytes[13], &bytes[12], &bytes[11], &bytes[10], &bytes[9], &bytes[8], &bytes[7], &bytes[6], &bytes[5], &bytes[4],
		&bytes[3], &bytes[2], &bytes[1], &bytes[0])
	if err != nil || n != 32 {
		return "", false
	}
	return ToIPv6(bytes), true
}

// ToIPv6 takes an ipv6 address of bytes and returns a string representation
func ToIPv6(in []byte) string {
	out := make([]string, len(in)+8) // 7 : characters
	for i, j := 0, 0; i < len(in); i, j = i+1, j+1 {
		out[j] = string(in[i])
		if i != len(in)-1 && (i+1)%4 == 0 {
			j++
			out[j] = ":"
		}
	}
	return strings.Join(out, "")
}

// FQDNTrim trims the trailing .
func FQDNTrim(name string) string {
	if dns.IsFqdn(name) {
		return strings.TrimRight(name, ".")
	}
	return name
}

// IsETLD returns true iff hostAddress is an etld
// amazon.co.uk == true
// sub.amazon.co.uk == false
func IsETLD(hostAddress string) bool {
	tld, err := GetETLD(hostAddress)
	if err != nil {
		return false
	}
	return hostAddress == tld
}

// GetETLD returns just the eltd of the supplied host address
func GetETLD(hostAddress string) (string, error) {
	// handle bug: https://github.com/golang/go/issues/20059
	special := SpecialETLD(hostAddress)
	if special != hostAddress {
		return special, nil
	}
	return publicsuffix.EffectiveTLDPlusOne(hostAddress)
}

// SpecialETLD case where publicsuffix doesn't fit what we
// want to test for etlds.
func SpecialETLD(hostAddress string) string {
	hostAddress = FQDNTrim(hostAddress)

	split := strings.Split(hostAddress, ".")
	if len(split) < 2 {
		return hostAddress
	}

	tld := strings.Join(split[len(split)-2:len(split)], ".")
	if _, ok := SpecialCaseTLDs[tld]; ok {
		return tld
	}

	return hostAddress
}

// SplitAddresses preserving etld. sub1.sub2.test.co.uk would become
// []string{"test.co.uk", "sub2.test.co.uk"}
// returns nil if hostAddress = eltd and does not return the original
// address that was supplied.
func SplitAddresses(hostAddress string) ([]string, error) {
	tld, err := GetETLD(hostAddress)
	if err != nil {
		return nil, err
	}
	addresses := make([]string, 0)
	addresses = append(addresses, tld)

	if hostAddress == tld {
		return nil, nil
	}

	subdomains := strings.TrimSuffix(hostAddress, "."+tld)
	subs := strings.Split(subdomains, ".")

	prev := "." + tld
	for i := len(subs) - 1; i > 0; i-- {
		subdomain := subs[i] + prev
		if subdomain == hostAddress {
			break
		}
		prev = "." + subdomain
		addresses = append(addresses, subdomain)
	}
	return addresses, nil
}
