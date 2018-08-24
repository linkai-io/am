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

	"golang.org/x/net/publicsuffix"
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
	ErrInvalidIP = errors.New("parsing IP address failure")
	// ErrHostMatchesETLD if the host is a public suffix
	ErrHostMatchesETLD = errors.New("host provided matches a top level domain")
	// ErrInvalidURLHostname url parses, but hostname is invalid
	ErrInvalidURLHostname = errors.New("hostname in url is invalid")
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
	var err error
	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)

	addresses := make(map[string]struct{}, 0)
	parserErrors := make([]*ParseError, 0)
	lineNo := 0

	replacer := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "", "*", "", "..", ".")

	for scanner.Scan() {
		lineNo++
		line := replacer.Replace(scanner.Text())

		if IsCIDR(line) {
			err = addCIDR(line, addresses, &parserErrors, lineNo, maxAddresses)
		} else if IsIP(line) {
			err = addIP(line, addresses, maxAddresses)
		} else if IsURL(line) {
			err = addURL(line, addresses, &parserErrors, lineNo, maxAddresses)
		} else {
			err = addHost(line, addresses, &parserErrors, lineNo, maxAddresses)
		}
		if err == ErrTooManyAddresses {
			parserErrors = append(parserErrors, &ParseError{LineNumber: lineNo, Line: line, Err: ErrTooManyAddresses})
			return addresses, parserErrors
		}
	}

	return addresses, parserErrors
}

func IsCIDR(line string) bool {
	if _, _, err := net.ParseCIDR(line); err != nil {
		return false
	}
	return true
}

func IsIP(line string) bool {
	if ip := net.ParseIP(line); ip == nil {
		return false
	}
	return true
}

func IsURL(line string) bool {
	if p, err := url.Parse(line); err != nil || p.Hostname() == "" {
		return false
	}
	return true
}

// ParseHost removes trailing/leading .
//
func ParseHost(host string) (string, error) {
	host = strings.Trim(host, ".")
	if !strings.Contains(host, ".") {
		return "", ErrMissingDot
	}

	// the string was all numbers but was not a valid ip,
	if match := ipMatch.MatchString(host); match {
		return "", ErrInvalidIP
	}

	suffix, _ := publicsuffix.PublicSuffix(host)
	if suffix == host {
		return "", ErrHostMatchesETLD
	}

	return host, nil
}

func addCIDR(line string, addresses map[string]struct{}, parserErrors *[]*ParseError, lineNo int, maxAddresses int) error {
	if ipAddress, ipNet, err := net.ParseCIDR(line); err == nil {
		ones, bits := ipNet.Mask.Size()
		useableAddresses := math.Pow(2, float64(bits)-float64(ones))
		if useableAddresses > float64(maxAddresses) {
			*parserErrors = append(*parserErrors, &ParseError{LineNumber: lineNo, Line: line, Err: ErrTooManyCIDRAddresses})
			return nil
		}

		for ipAddress := ipAddress.Mask(ipNet.Mask); ipNet.Contains(ipAddress); inc(ipAddress) {
			if err := addIP(ipAddress.String(), addresses, maxAddresses); err != nil {
				return err
			}
		}
	}
	return nil
}

func addURL(line string, addresses map[string]struct{}, parserErrors *[]*ParseError, lineNo int, maxAddresses int) error {
	p, _ := url.Parse(line)
	if IsIP(p.Hostname()) {
		return addIP(p.Hostname(), addresses, maxAddresses)
	}

	return addHost(p.Hostname(), addresses, parserErrors, lineNo, maxAddresses)
}

func addIP(address string, addresses map[string]struct{}, maxAddresses int) error {
	if len(addresses) >= maxAddresses {
		return ErrTooManyAddresses
	}
	addresses[address] = struct{}{}
	return nil
}

func addHost(line string, addresses map[string]struct{}, parserErrors *[]*ParseError, lineNo int, maxAddresses int) error {
	if len(addresses) >= maxAddresses {
		return ErrTooManyAddresses
	}

	host, err := ParseHost(line)
	if err != nil {
		*parserErrors = append(*parserErrors, &ParseError{LineNumber: lineNo, Line: line, Err: err})
		return nil
	}

	addresses[host] = struct{}{}
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
