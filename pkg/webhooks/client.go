package webhooks

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/linkai-io/am/pkg/parsers"
)

// Client handles the transport of sending the events over an HTTPS connection
type Client struct {
	c *http.Client
}

// NewClient creates a secure HTTPS client for sending webhook events
func NewClient() *Client {
	timeout := 10 * time.Second
	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			c, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
			if parsers.IsBannedIP(ip) {
				log.Printf("BANNED IP")
				return nil, errors.New("ip address is banned")
			}
			return c, err
		},
		DialTLS: func(network, addr string) (net.Conn, error) {
			c, err := tls.Dial(network, addr, &tls.Config{})
			if err != nil {
				return nil, err
			}

			ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
			if parsers.IsBannedIP(ip) {
				log.Printf("TLS BANNED IP")
				return nil, errors.New("ip address is banned")
			}

			err = c.Handshake()
			if err != nil {
				return c, err
			}

			return c, c.Handshake()
		},
		TLSHandshakeTimeout: 9 * time.Second,
	}

	return &Client{c: &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: tr,
		Timeout:   timeout,
	}}
}

// SendEvent depending on type (slack/custom/custom_signed etc)
func (c *Client) SendEvent(ctx context.Context, evt *Data) (int, error) {
	switch evt.Settings.Type {
	case "slack":
		return c.sendSlackEvent(ctx, evt)
	case "custom":
		return c.sendCustomEvent(ctx, evt)
	case "custom_signed":
		return c.sendCustomSignedEvent(ctx, evt)
	}
	return 0, errors.New("invalid webhook type")
}

func (c *Client) sendSlackEvent(ctx context.Context, evt *Data) (int, error) {
	msg, err := FormatSlackMessage(evt.Settings.ScanGroupName, evt)
	if err != nil {
		return 0, err
	}
	resp, err := c.c.Post(evt.Settings.URL, "application/json", strings.NewReader(msg))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *Client) sendCustomEvent(ctx context.Context, evt *Data) (int, error) {
	return 0, nil
}

func (c *Client) sendCustomSignedEvent(ctx context.Context, evt *Data) (int, error) {
	return 0, nil
}
