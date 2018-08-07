package inputlist

import (
	"bufio"
	"errors"
	"io"
	"math"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// TODO: enhance invalid ipv6 detection
var ipMatch = regexp.MustCompile("^([0-9].*[0-9])$")

var (
	// ErrMissingDot If a line contains no period (and not an ipv6 address)
	ErrMissingDot = errors.New("missing period")
	// ErrTooManyCIDRAddresses If too many addresses in a CIDR block are encountered
	ErrTooManyCIDRAddresses = errors.New("too many addresses in CIDR block")
	// ErrTooManyAddresses if we exceeded our total address count
	ErrTooManyAddresses = errors.New("too many addresses in input list")
	// ErrInvalidIP if the ip address is malformed (starts/ends with number but does not parse)
	ErrInvalidIP = errors.New("error parsing IP address")
)

// ParseError contains the line number, line and parse error
type ParseError struct {
	LineNumber int
	Line       string
	Err        error
}

// ParseList parses a list of hostnames, domains, urls, ip addresses, and cidr ranges
// and returns a de-duplicated list of strings and any errors with line numbers
// returns nil if number of addresses exceeds maxAddress
func ParseList(in io.Reader, maxAddresses int) (map[string]struct{}, []*ParseError) {
	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)

	addresses := make(map[string]struct{}, 0)
	errors := make([]*ParseError, 0)
	idx := 0

	replacer := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "", "*", "", "..", ".")

	for scanner.Scan() {
		idx++
		line := replacer.Replace(scanner.Text())

		if ipAddress, ipNet, err := net.ParseCIDR(line); err == nil {
			ones, bits := ipNet.Mask.Size()
			useableAddresss := math.Pow(2, float64(bits)-float64(ones))
			if useableAddresss > float64(maxAddresses) {
				errors = append(errors, &ParseError{LineNumber: idx, Line: line, Err: ErrTooManyCIDRAddresses})
				continue
			}

			for ipAddress := ipAddress.Mask(ipNet.Mask); ipNet.Contains(ipAddress); inc(ipAddress) {
				if err := addAddress(ipAddress.String(), addresses, maxAddresses); err != nil {
					errors = append(errors, &ParseError{LineNumber: idx, Line: line, Err: ErrTooManyAddresses})
					return nil, errors
				}
			}
			continue
		}

		if p, err := url.Parse(line); err == nil && p.Hostname() != "" {
			if err := addAddress(p.Hostname(), addresses, maxAddresses); err != nil {
				errors = append(errors, &ParseError{LineNumber: idx, Line: line, Err: err})
				if err == ErrTooManyAddresses {
					return nil, errors
				}
			}
			continue
		}

		if !strings.Contains(line, ".") && !strings.Contains(line, ":") {
			errors = append(errors, &ParseError{LineNumber: idx, Line: line, Err: ErrMissingDot})
			continue
		}

		if err := addAddress(line, addresses, maxAddresses); err != nil {
			errors = append(errors, &ParseError{LineNumber: idx, Line: line, Err: err})
			if err == ErrTooManyAddresses {
				return nil, errors
			}
		}
	}
	return addresses, errors
}

func addAddress(address string, addresses map[string]struct{}, maxAddresses int) error {
	if len(addresses) >= maxAddresses {
		return ErrTooManyAddresses
	}

	// remove period at start of string if there is one
	if strings.Index(address, ".") == 0 {
		address = address[1:]
	}

	// remove period at end of string if there is one
	if address[len(address)-1:len(address)] == "." {
		address = address[:len(address)-1]
	}

	// if starts with a number and ends with a number, check if it's a valid IP
	match := ipMatch.MatchString(address)
	if match && net.ParseIP(address) == nil {
		return ErrInvalidIP
	}

	addresses[address] = struct{}{}
	return nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
