package am

var (
	EventInitialGroupCompleteID int32 = 1
	EventMaxHostPricingID       int32 = 2
	EventNewHostID              int32 = 10
	EventNewRecordID            int32 = 11
	EventNewOpenPortID          int32 = 12
	EventClosedPortID           int32 = 13
	EventNewWebsiteID           int32 = 100
	EventWebHTMLUpdatedID       int32 = 101
	EventNewWebTechID           int32 = 102
	EventWebJSChangedID         int32 = 103
	EventCertExpiringID         int32 = 150
	EventCertExpiredID          int32 = 151
	EventAXFRID                 int32 = 200
	EventNSECID                 int32 = 201
)
var EventTypes = map[int32]string{
	1:   "initial scan group analysis completed",
	2:   "maximum number of hostnames reached for pricing plan",
	10:  "new hostname",
	11:  "new record",
	12:  "new port open",
	13:  "port closed",
	100: "new website detected",
	101: "website's html updated",
	102: "website's technology changed or updated",
	103: "website's javascript changed",
	150: "certificate expiring",
	151: "certificate expired",
	200: "dns server exposing records via zone transfer",
	201: "dns server exposing records via NSEC walking",
}

type EventInitialGroupComplete struct {
	Message string `json:"message"`
}

type EventNewHost struct {
	Host string `json:"new_host"`
}

type EventNewOpenPort struct {
	Host       string  `json:"hostname"`
	CurrentIP  string  `json:"current_ip"`
	PreviousIP string  `json:"previous_ip"`
	OpenPorts  []int32 `json:"open_ports"`
}

type EventClosedPort struct {
	Host        string  `json:"hostname"`
	CurrentIP   string  `json:"current_ip"`
	PreviousIP  string  `json:"previous_ip"`
	ClosedPorts []int32 `json:"closed_ports"`
}

type EventNewWebsite struct {
	LoadURL string `json:"load_url"`
	URL     string `json:"url"`
	Port    int    `json:"port"`
}

type EventNewWebTech struct {
	LoadURL  string `json:"load_url"`
	URL      string `json:"url"`
	Port     int    `json:"port"`
	TechName string `json:"tech_name"`
	Version  string `json:"tech_version"`
}

type EventCertExpiring struct {
	SubjectName   string `json:"subject_name"`
	Port          int    `json:"port"`
	ValidTo       int64  `json:"valid_to"`
	TimeRemaining string `json:"time_remaining"`
}

type EventCertExpired struct {
	SubjectName string `json:"subject_name"`
	Port        int    `json:"port"`
	ValidTo     int64  `json:"valid_to"`
}

type EventAXFR struct {
	Servers []string `json:"servers"`
}

type EventNSEC struct {
	Servers []string `json:"servers"`
}
