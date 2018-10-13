package am

// HTTPResponse represents a captured network response
type HTTPResponse struct {
	Address    *ScanGroupAddress
	Port       string
	RequestID  string
	Status     int
	StatusText string
	URL        string
	Headers    map[string]interface{}
	MimeType   string
	Body       string
}
