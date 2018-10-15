package am

// HTTPResponse represents a captured network response
type HTTPResponse struct {
	Scheme            string
	Host              string
	ResponsePort      string
	RequestedPort     string
	RequestID         string
	Status            int
	StatusText        string
	URL               string
	Headers           map[string]interface{}
	MimeType          string
	Body              string
	ResponseTimestamp int64
	IsDocument        bool
	WebCertificate    *WebCertificate
}

type WebCertificate struct {
	Protocol                          string   `json:"protocol"`                          // Protocol name (e.g. "TLS 1.2" or "QUIC").
	KeyExchange                       string   `json:"keyExchange"`                       // Key Exchange used by the connection, or the empty string if not applicable.
	KeyExchangeGroup                  string   `json:"keyExchangeGroup,omitempty"`        // (EC)DH group used by the connection, if applicable.
	Cipher                            string   `json:"cipher"`                            // Cipher name.
	Mac                               string   `json:"mac,omitempty"`                     // TLS MAC. Note that AEAD ciphers do not have separate MACs.
	CertificateId                     int      `json:"certificateId"`                     // Certificate ID value.
	SubjectName                       string   `json:"subjectName"`                       // Certificate subject name.
	SanList                           []string `json:"sanList"`                           // Subject Alternative Name (SAN) DNS names and IP addresses.
	Issuer                            string   `json:"issuer"`                            // Name of the issuing CA.
	ValidFrom                         float64  `json:"validFrom"`                         // Certificate valid from date.
	ValidTo                           float64  `json:"validTo"`                           // Certificate valid to (expiration) date
	CertificateTransparencyCompliance string   `json:"certificateTransparencyCompliance"` // Whether the request complied with Certificate Transparency policy enum values: unknown, not-compliant, compliant
}

type WebData struct {
	Address       *ScanGroupAddress
	Responses     []*HTTPResponse
	Image         string
	SerializedDOM string
}
