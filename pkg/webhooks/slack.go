package webhooks

import (
	"encoding/json"
	"fmt"
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

	for _, e := range evt.Event {
		var alertText string
		msg := ""
		switch e.TypeID {
		case am.EventCertExpiredID:
		case am.EventCertExpiringID:
			alertText = "*The following certificates will expire soon:*"

			if e.JSONData != "" && e.JSONData != "{}" {
				// handle new json type
				var expireCerts []*am.EventCertExpiring
				if err := json.Unmarshal([]byte(e.JSONData), &expireCerts); err != nil {
					return "", err
				}

				for _, expired := range expireCerts {
					msg += fmt.Sprintf("• %s on port %d expires in %s\n", expired.SubjectName, expired.Port, FormatUnixTimeRemaining(expired.ValidTo))
				}
			}

		case am.EventNewOpenPortID:
			alertText = "*The following ports were opened:*"

			if e.JSONData != "" && e.JSONData != "{}" {
				var openPorts []*am.EventNewOpenPort
				if err := json.Unmarshal([]byte(e.JSONData), &openPorts); err != nil {
					return "", err
				}

				for _, open := range openPorts {
					ips := open.CurrentIP
					if open.CurrentIP != open.PreviousIP {
						ips += ") previously (" + open.PreviousIP
					}
					msg += fmt.Sprintf("• Host %s (%s) ports: %s\n", open.Host, ips, IntToString(open.OpenPorts))
				}
			}
		case am.EventClosedPortID:
			alertText = "*The following ports were recently closed:*"
			if e.JSONData != "" && e.JSONData != "{}" {
				var closedPorts []*am.EventClosedPort
				if err := json.Unmarshal([]byte(e.JSONData), &closedPorts); err != nil {
					return "", err
				}

				for _, closed := range closedPorts {
					ips := closed.CurrentIP
					if closed.CurrentIP != closed.PreviousIP {
						ips += ") previously (" + closed.PreviousIP
					}
					msg += fmt.Sprintf("• Host %s (%s) ports: %s\n", closed.Host, ips, IntToString(closed.ClosedPorts))
				}
			}
		case am.EventInitialGroupCompleteID:
		case am.EventMaxHostPricingID:
		case am.EventNewHostID:
			alertText = "*The following new hosts were found:*"

			if e.JSONData != "" && e.JSONData != "{}" {
				var newHosts []*am.EventNewHost
				if err := json.Unmarshal([]byte(e.JSONData), &newHosts); err != nil {
					return "", err
				}
				msg += "• "
				for i, newHost := range newHosts {
					msg += newHost.Host
					if i != len(newHosts)-1 {
						msg += ","
					}
				}
			}

		case am.EventAXFRID:
			alertText = "*The following name servers allow zone transfers (AXFR):*"

			if e.JSONData != "" && e.JSONData != "{}" {
				var axfrServers []*am.EventAXFR
				if err := json.Unmarshal([]byte(e.JSONData), &axfrServers); err != nil {
					return "", err
				}

				for _, axfr := range axfrServers {
					msg += "• " + strings.Join(axfr.Servers, ",") + "\n"
				}
			}
		case am.EventNSECID:
			alertText = "*The following name servers are leaking hostnames via NSEC records:*"

			if e.JSONData != "" && e.JSONData != "{}" {
				var nsecServers []*am.EventNSEC
				if err := json.Unmarshal([]byte(e.JSONData), &nsecServers); err != nil {
					return "", err
				}

				for _, nsec := range nsecServers {
					msg += "• " + strings.Join(nsec.Servers, ",") + "\n"
				}
			}

		case am.EventNewWebsiteID:
			alertText = "*The following new web sites were found:*"

			if e.JSONData != "" && e.JSONData != "{}" {
				var newSites []*am.EventNewWebsite
				if err := json.Unmarshal([]byte(e.JSONData), &newSites); err != nil {
					return "", err
				}

				for _, site := range newSites {
					if wasRedirected(site.LoadURL, site.URL) {
						msg += fmt.Sprintf("• %s (was redirected to %s) on port %d\n", site.LoadURL, site.URL, site.Port)
					} else {
						msg += fmt.Sprintf("• %s on port %d\n", site.LoadURL, site.Port)
					}
				}
			}
		case am.EventWebHTMLUpdatedID:
		case am.EventWebJSChangedID:
		case am.EventNewWebTechID:
			alertText = "*The following new or updated technologies were found:*"

			if e.JSONData != "" && e.JSONData != "{}" {
				var newTech []*am.EventNewWebTech
				if err := json.Unmarshal([]byte(e.JSONData), &newTech); err != nil {
					return "", err
				}

				for _, tech := range newTech {
					if wasRedirected(tech.LoadURL, tech.URL) {
						msg += fmt.Sprintf("• %s (was redirected to %s) is running %s %s\n", tech.LoadURL, tech.URL, tech.TechName, tech.Version)
					} else {
						msg += fmt.Sprintf("• %s is running %s %s\n", tech.LoadURL, tech.TechName, tech.Version)
					}
				}
			}
		}
		if msg == "" {
			continue
		}
		alert := NewSlackSection(alertText)
		//alert.AppendFields(alertFields)
		blocks.AppendSection(alert)
		data := NewSlackSection(msg)
		blocks.AppendSection(data)
	}

	ts := time.Unix(0, evt.Event[0].EventTimestamp)
	footer := NewSlackSection(fmt.Sprintf("<!date^%d^ Events occurred at: {date_num} {time_secs}|%s>, <https://console.linkai.io/login/|login> to view", ts.Unix(), ts))
	blocks.AppendSection(footer)
	data, err := json.Marshal(blocks)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
