package webhooks

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/linkai-io/am/am"
)

type SlackImageTitle struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type SlackImage struct {
	Type     string           `json:"type"`
	Title    *SlackImageTitle `json:"title"`
	ImageURL string           `json:"image_url"`
	AltText  string           `json:"alt_text"`
}

type SlackField struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Emoji bool   `json:"emoji,omitempty"`
}

func NewSlackField(text string) *SlackField {
	return &SlackField{Type: "mrkdwn", Text: text}
}

func NewSlackFieldDivider() *SlackField {
	return &SlackField{Type: "divider"}
}

type SlackAccessory struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url"`
	AltText  string `json:"alt_text"`
}

type SlackSection struct {
	Type      string          `json:"type"`
	Text      *SlackField     `json:"text"`
	Fields    []*SlackField   `json:"fields,omitempty"`
	Accessory *SlackAccessory `json:"accessory,omitempty"`
}

func NewSlackSection(text string) *SlackSection {
	return &SlackSection{Type: "section", Text: &SlackField{Type: "mrkdwn", Text: text}}
}

func (s *SlackSection) AppendFields(field []*SlackField) {
	if s.Fields == nil {
		s.Fields = make([]*SlackField, 0)
	}
	s.Fields = append(s.Fields, field...)
}

func (s *SlackSection) AppendField(field *SlackField) {
	if s.Fields == nil {
		s.Fields = make([]*SlackField, 0)
	}
	s.Fields = append(s.Fields, field)
}

type SlackBlock struct {
	Blocks []*SlackSection `json:"blocks"`
}

func (b *SlackBlock) AppendSection(section *SlackSection) {
	if b.Blocks == nil {
		b.Blocks = make([]*SlackSection, 0)
	}
	b.Blocks = append(b.Blocks, section)
}

func FormatSlackMessage(groupName string, evt *Data) (string, error) {
	blocks := &SlackBlock{}
	title := &SlackSection{Type: "section",
		Text: &SlackField{Type: "mrkdwn", Text: ":information_source: *Hakken Alert (" + groupName + ")*"},
		//Accessory: &SlackAccessory{Type: "image", ImageURL: "https://linkai.io/images/logo-0.5.png", AltText: "linkai logo"},
	}
	blocks.AppendSection(title)

	var alertText string
	var alertFields []*SlackField
	for _, e := range evt.Event {
		switch e.TypeID {
		case am.EventCertExpiredID:
		case am.EventCertExpiringID:
			alertText = "The following certificates will expire soon:"

			if e.JSONData != "" && e.JSONData != "{}" {
				// handle new json type
				var expireCerts []*am.EventCertExpiring
				if err := json.Unmarshal([]byte(e.JSONData), &expireCerts); err != nil {
					return "", err
				}

				alertFields = make([]*SlackField, len(expireCerts))
				for i, expired := range expireCerts {
					alertFields[i] = NewSlackField(fmt.Sprintf("%s on port %d expires in %s\n", expired.SubjectName, expired.Port, FormatUnixTimeRemaining(expired.ValidTo)))
				}

			}

		case am.EventNewOpenPortID:
			alertText = "The following ports were opened:"

			if e.JSONData != "" && e.JSONData != "{}" {
				var openPorts []*am.EventNewOpenPort
				if err := json.Unmarshal([]byte(e.JSONData), &openPorts); err != nil {
					return "", err
				}

				alertFields = make([]*SlackField, len(openPorts))
				for i, open := range openPorts {
					ips := open.CurrentIP
					if open.CurrentIP != open.PreviousIP {
						ips += ") previously (" + open.PreviousIP
					}
					alertFields[i] = NewSlackField(fmt.Sprintf("Host %s (%s) ports: %s\n", open.Host, ips, IntToString(open.OpenPorts)))
				}
			}
		case am.EventClosedPortID:
			alertText = "The following ports were recently closed:"
			if e.JSONData != "" && e.JSONData != "{}" {
				var closedPorts []*am.EventClosedPort
				if err := json.Unmarshal([]byte(e.JSONData), &closedPorts); err != nil {
					return "", err
				}

				alertFields = make([]*SlackField, len(closedPorts))
				for i, closed := range closedPorts {
					ips := closed.CurrentIP
					if closed.CurrentIP != closed.PreviousIP {
						ips += ") previously (" + closed.PreviousIP
					}
					alertFields[i] = NewSlackField(fmt.Sprintf("Host %s (%s) ports: %s\n", closed.Host, ips, IntToString(closed.ClosedPorts)))
				}
			}
		case am.EventInitialGroupCompleteID:
		case am.EventMaxHostPricingID:
		case am.EventNewHostID:
			alertText = "The following new hosts were found:"

			if e.JSONData != "" && e.JSONData != "{}" {
				var newHosts []*am.EventNewHost
				if err := json.Unmarshal([]byte(e.JSONData), &newHosts); err != nil {
					return "", err
				}
				alertFields = make([]*SlackField, len(newHosts))
				for i, newHost := range newHosts {
					alertFields[i] = NewSlackField(newHost.Host)
				}
			}

		case am.EventAXFRID:
			alertText = "The following name servers allow zone transfers (AXFR):"

			if e.JSONData != "" && e.JSONData != "{}" {
				var axfrServers []*am.EventAXFR
				if err := json.Unmarshal([]byte(e.JSONData), &axfrServers); err != nil {
					return "", err
				}
				alertFields = make([]*SlackField, len(axfrServers))
				for i, axfr := range axfrServers {
					log.Printf("%#v %s", axfr, strings.Join(axfr.Servers, ","))
					alertFields[i] = NewSlackField(strings.Join(axfr.Servers, ","))
				}
			}
		case am.EventNSECID:
			alertText = "The following name servers are leaking hostnames via NSEC records:"

			if e.JSONData != "" && e.JSONData != "{}" {
				var nsecServers []*am.EventNSEC
				if err := json.Unmarshal([]byte(e.JSONData), &nsecServers); err != nil {
					return "", err
				}
				alertFields = make([]*SlackField, len(nsecServers))
				for i, nsec := range nsecServers {
					alertFields[i] = NewSlackField(strings.Join(nsec.Servers, ","))
				}
			}

		case am.EventNewWebsiteID:
			alertText = "The following new web sites were found:"

			if e.JSONData != "" && e.JSONData != "{}" {
				var newSites []*am.EventNewWebsite
				if err := json.Unmarshal([]byte(e.JSONData), &newSites); err != nil {
					return "", err
				}

				alertFields = make([]*SlackField, len(newSites))
				for i, site := range newSites {
					msg := ""
					if wasRedirected(site.LoadURL, site.URL) {
						msg = fmt.Sprintf("%s (was redirected to %s) on port %d", site.LoadURL, site.URL, site.Port)
					} else {
						msg = fmt.Sprintf("%s on port %d", site.LoadURL, site.Port)
					}
					alertFields[i] = NewSlackField(msg)
				}
			}
		case am.EventWebHTMLUpdatedID:
		case am.EventWebJSChangedID:
		case am.EventNewWebTechID:
			alertText = "The following new or updated technologies were found:"

			if e.JSONData != "" && e.JSONData != "{}" {
				var newTech []*am.EventNewWebTech
				if err := json.Unmarshal([]byte(e.JSONData), &newTech); err != nil {
					return "", err
				}

				alertFields = make([]*SlackField, len(newTech))
				for i, tech := range newTech {
					msg := ""
					if wasRedirected(tech.LoadURL, tech.URL) {
						msg = fmt.Sprintf("%s (was redirected to %s) is running %s %s", tech.LoadURL, tech.URL, tech.TechName, tech.Version)
					} else {
						msg = fmt.Sprintf("%s is running %s %s", tech.LoadURL, tech.TechName, tech.Version)
					}
					alertFields[i] = NewSlackField(msg)
				}
			}
		}
	}

	alert := NewSlackSection(alertText)
	alert.AppendFields(alertFields)
	blocks.AppendSection(alert)
	ts := time.Unix(0, evt.Event[0].EventTimestamp)
	footer := NewSlackSection(fmt.Sprintf("<!date^%d^ Event occurred at: {date_num} {time_secs}|%s>, <https://console.linkai.io/login/|login> to view", ts.Unix(), ts))
	blocks.AppendSection(footer)
	data, err := json.Marshal(blocks)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
