package certstream

// CertStreamEntry from the certstream service
type CertStreamEntry struct {
	MessageType string `json:"message_type"`
	Data        struct {
		UpdateType string `json:"update_type"`
		LeafCert   struct {
			Subject struct {
				Aggregated string      `json:"aggregated"`
				C          interface{} `json:"C"`
				ST         interface{} `json:"ST"`
				L          interface{} `json:"L"`
				O          interface{} `json:"O"`
				OU         interface{} `json:"OU"`
				CN         string      `json:"CN"`
			} `json:"subject"`
			Extensions struct {
				KeyUsage               string `json:"keyUsage"`
				ExtendedKeyUsage       string `json:"extendedKeyUsage"`
				BasicConstraints       string `json:"basicConstraints"`
				SubjectKeyIdentifier   string `json:"subjectKeyIdentifier"`
				AuthorityKeyIdentifier string `json:"authorityKeyIdentifier"`
				AuthorityInfoAccess    string `json:"authorityInfoAccess"`
				SubjectAltName         string `json:"subjectAltName"`
				CertificatePolicies    string `json:"certificatePolicies"`
			} `json:"extensions"`
			NotBefore    float64  `json:"not_before"`
			NotAfter     float64  `json:"not_after"`
			SerialNumber string   `json:"serial_number"`
			Fingerprint  string   `json:"fingerprint"`
			AsDer        string   `json:"as_der"`
			AllDomains   []string `json:"all_domains"`
		} `json:"leaf_cert"`
		Chain []struct {
			Subject struct {
				Aggregated string      `json:"aggregated"`
				C          string      `json:"C"`
				ST         interface{} `json:"ST"`
				L          interface{} `json:"L"`
				O          string      `json:"O"`
				OU         interface{} `json:"OU"`
				CN         string      `json:"CN"`
			} `json:"subject"`
			Extensions struct {
				BasicConstraints       string `json:"basicConstraints,omitempty"`
				KeyUsage               string `json:"keyUsage,omitempty"`
				AuthorityInfoAccess    string `json:"authorityInfoAccess,omitempty"`
				AuthorityKeyIdentifier string `json:"authorityKeyIdentifier,omitempty"`
				CertificatePolicies    string `json:"certificatePolicies,omitempty"`
				CrlDistributionPoints  string `json:"crlDistributionPoints,omitempty"`
				SubjectKeyIdentifier   string `json:"subjectKeyIdentifier,omitempty"`
			} `json:"extensions,omitempty"`
			NotBefore    float64 `json:"not_before"`
			NotAfter     float64 `json:"not_after"`
			SerialNumber string  `json:"serial_number"`
			Fingerprint  string  `json:"fingerprint"`
			AsDer        string  `json:"as_der"`
		} `json:"chain"`
		CertIndex int     `json:"cert_index"`
		Seen      float64 `json:"seen"`
		Source    struct {
			URL  string `json:"url"`
			Name string `json:"name"`
		} `json:"source"`
	} `json:"data"`
}
