package parsers

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/net/publicsuffix"
)

var (
	ErrHostIsIPAddress = errors.New("provided host is an ip address")
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
	hostAddress = strings.ToLower(hostAddress)
	tld, err := GetETLD(hostAddress)
	if err != nil {
		return false
	}
	return hostAddress == tld
}

// GetETLD returns just the eltd of the supplied host address
func GetETLD(hostAddress string) (string, error) {
	if ip := net.ParseIP(hostAddress); ip != nil {
		return "", ErrHostIsIPAddress
	}

	hostAddress = strings.ToLower(hostAddress)
	// handle bug: https://github.com/golang/go/issues/20059
	if _, ok := SpecialCaseTLDs[hostAddress]; ok {
		return hostAddress, nil
	}
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
	hostAddress = strings.ToLower(hostAddress)

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
	hostAddress = strings.ToLower(hostAddress)
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

// GetDepth returns how many subdomains the host address has:
// ex: test1.test2.example.com would return 3.
// ex2: test2.example.com would return 2
// ex3: test1.amazon.co.uk would return 2
func GetDepth(hostAddress string) (int, error) {
	hostAddress = strings.ToLower(hostAddress)
	tld, err := GetETLD(hostAddress)
	if err != nil {
		return 0, err
	}

	if hostAddress == tld {
		return 1, nil
	}

	subdomains := strings.TrimSuffix(hostAddress, "."+tld)
	subs := strings.Split(subdomains, ".")
	return len(subs) + 1, nil // 1 for tld
}

// GetSubDomain returns the last sub domain part of a host address
func GetSubDomain(hostAddress string) (string, error) {
	hostAddress = strings.ToLower(hostAddress)
	sub, _, err := GetSubDomainAndDomain(hostAddress)
	return sub, err
}

// GetSubDomainAndDomain returns the subdomain + the rest of the domain, or error
func GetSubDomainAndDomain(hostAddress string) (string, string, error) {
	hostAddress = strings.ToLower(hostAddress)
	tld, err := GetETLD(hostAddress)
	if err != nil {
		return "", "", err
	}

	if hostAddress == tld {
		return "", hostAddress, nil
	}

	subdomains := strings.TrimSuffix(hostAddress, "."+tld)
	subs := strings.Split(subdomains, ".")
	if len(subs) == 1 {
		return subs[0], tld, nil
	}
	domain := strings.TrimLeft(strings.Join(subs[1:], ".")+"."+tld, ".")
	return subs[0], domain, nil
}

// ExtractHostsFromResponses returns potential hosts from a response
func ExtractHostsFromResponses(needles []*regexp.Regexp, haystacks []string) map[string]struct{} {
	hosts := make(map[string]struct{}, 0)
	for _, haystack := range haystacks {
		found := ExtractHostsFromResponse(needles, haystack)
		if len(found) > 0 {
			for k, v := range found {
				hosts[k] = v
			}
		}
	}
	return hosts
}

// ExtractHostsFromResponse returns potential hosts from a response
// TODO: make this a bit more robust.
func ExtractHostsFromResponse(needles []*regexp.Regexp, haystack string) map[string]struct{} {

	hosts := make(map[string]struct{}, 0)
	for _, needle := range needles {
		indexes := needle.FindAllStringIndex(haystack, -1)
		if len(indexes) == 0 {
			continue
		}

		for _, index := range indexes {

			if len(index) != 2 {
				continue
			}

			prefixLen := 50
			matchLen := index[1] - index[0]
			start := index[0] - prefixLen

			if index[0] < prefixLen {
				prefixLen = index[0]
				start = index[0] - prefixLen
			}
			end := index[1]

			match := haystack[start:end]

			// walk backwards from our match and stop when we find a non-valid character
			for i := prefixLen - 1; i >= 0; i-- {
				switch match[i] {
				case '<', '>', '/', '\\', '\'',
					'"', ':', '_', '=', '*',
					' ', '\t', '\r', '\n':
					start = start + i + 1
					goto FOUND
				case '%':
					start = start + i + 3 // remove hex
					goto FOUND
				}
			}
		FOUND:
			// make sure we don't go past the actual size of the domain via the prefixlen loop
			if len(haystack[start:end]) < matchLen {
				continue
			}
			//log.Printf("FOUND: %s\norig: %s\n", haystack[start:end], haystack[index[0]-30:index[1]])
			newHost := strings.ToLower(strings.TrimLeft(haystack[start:end], "."))
			hosts[newHost] = struct{}{}
		}
	}
	return hosts
}
