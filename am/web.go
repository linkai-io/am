package am

const (
	RNWebData             = "lrn:service:webdata:feature:"
	RNWebDataResponses    = "lrn:service:webdata:feature:responses"
	RNWebDataCertificates = "lrn:service:webdata:feature:certificates"
	RNWebDataSnapshots    = "lrn:service:webdata:feature:snapshots"
	WebDataServiceKey     = "webdataservice"
)

// HTTPResponse represents a captured network response
type HTTPResponse struct {
	Scheme            string                 `json:"scheme"`
	Host              string                 `json:"host"`
	ResponsePort      string                 `json:"response_port"`
	RequestedPort     string                 `json:"requested_port"`
	RequestID         string                 `json:"request_id,omitempty"`
	Status            int                    `json:"status"`
	StatusText        string                 `json:"status_text"`
	URL               string                 `json:"url"`
	Headers           map[string]interface{} `json:"headers"`
	MimeType          string                 `json:"mime_type"`
	RawBody           string                 `json:"raw_body,omitempty"`
	RawBodyLink       string                 `json:"raw_body_link"`
	RawBodyHash       string                 `json:"raw_body_hash"`
	ResponseTimestamp int64                  `json:"response_timestamp"`
	IsDocument        bool                   `json:"is_document"`
	WebCertificate    *WebCertificate        `json:"web_certificate"`
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
	ValidFrom                         int64    `json:"validFrom"`                         // Certificate valid from date.
	ValidTo                           int64    `json:"validTo"`                           // Certificate valid to (expiration) date
	CertificateTransparencyCompliance string   `json:"certificateTransparencyCompliance"` // Whether the request complied with Certificate Transparency policy enum values: unknown, not-compliant, compliant
}

type WebData struct {
	Address           *ScanGroupAddress `json:"address"`
	Responses         []*HTTPResponse   `json:"responses"`
	Snapshot          string            `json:"snapshot,omitempty"`
	SnapshotLink      string            `json:"snapshot_link"`
	SerializedDOM     string            `json:"serialized_dom,omitempty"`
	SerializedDOMLink string            `json:"serialized_dom_link"`
	ResponseTimestamp int64             `json:"response_timeestamp"`
}
