package webflowclient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"regexp"
	"strings"
	"time"

	"github.com/linkai-io/am/pkg/parsers"

	"github.com/gammazero/workerpool"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/filestorage"
	"github.com/rs/zerolog/log"
)

type RequestEvent struct {
	UserContext am.UserContext          `json:"user_context"`
	Host        string                  `json:"host"`
	Ports       []int32                 `json:"ports"`
	Config      *am.CustomRequestConfig `json:"config"`
}

type Results struct {
	Results []*am.CustomWebFlowResults `json:"results"`
}

type Client struct {
	client  *http.Client
	storage filestorage.Storage
}

func New(store filestorage.Storage) *Client {
	timeout := time.Duration(15 * time.Second)

	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			c, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
			if parsers.IsBannedIP(ip) {
				return nil, errors.New("ip address is banned")
			}
			return c, err
		},
		DialTLS: func(network, addr string) (net.Conn, error) {
			c, err := tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: true})
			if err != nil {
				return nil, err
			}

			ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
			if parsers.IsBannedIP(ip) {
				return nil, errors.New("ip address is banned")
			}

			err = c.Handshake()
			if err != nil {
				return c, err
			}

			return c, c.Handshake()
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSHandshakeTimeout:   5 * time.Second,
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       0,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: timeout,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
	c := &Client{}
	c.client = client
	c.storage = store
	return c
}

func (c *Client) Do(ctx context.Context, event *RequestEvent) (*Results, error) {
	ports := event.Ports
	schemes := []string{"http", "https"}

	if event.Config.OnlyPort != 0 {
		ports = []int32{event.Config.OnlyPort}
	}

	if event.Config.OnlyScheme != "" {
		schemes = []string{event.Config.OnlyScheme}
	}

	if len(ports) == 0 {
		return nil, errors.New("no ports specified")
	}

	if parsers.IsBannedIP(event.Host) {
		return nil, errors.New("banned host")
	}
	if event.Host == "" {
		return nil, errors.New("no host specified")
	}

	return c.makeRequests(ctx, event, ports, schemes)
}

func (c *Client) makeRequests(ctx context.Context, event *RequestEvent, ports []int32, schemes []string) (*Results, error) {
	requestLen := len(ports) * len(schemes)
	results := make(chan *am.CustomWebFlowResults, requestLen)
	pool := workerpool.New(requestLen)

	for _, scheme := range schemes {
		for _, port := range ports {
			pool.Submit(c.processRequestTask(ctx, event, results, port, scheme))
		}
	}

	pool.StopWait()
	close(results)

	allResults := &Results{}
	allResults.Results = make([]*am.CustomWebFlowResults, 0)
	for result := range results {
		allResults.Results = append(allResults.Results, result)
	}
	return allResults, nil
}

func (c *Client) processRequestTask(ctx context.Context, event *RequestEvent, results chan<- *am.CustomWebFlowResults, port int32, scheme string) func() {
	return func() {
		url := fmt.Sprintf("%s://%s:%d%s", scheme, event.Host, port, event.Config.URI)
		result := &am.CustomWebFlowResults{}
		result.Result = make([]*am.CustomRequestResult, 0)
		result.LoadHostAddress = event.Host
		result.RequestedPort = port
		result.ResponsePort = port
		result.LoadURL = url
		result.URL = url

		req, err := buildRequest(event, url)
		if err != nil {
			handleError(ctx, result, results, event, port, scheme, url, err)
			return
		}

		trace := &httptrace.ClientTrace{
			GotConn: func(connInfo httptrace.GotConnInfo) {
				ip, _, _ := net.SplitHostPort(connInfo.Conn.RemoteAddr().String())
				result.LoadIPAddress = ip
				if parsers.IsBannedIP(ip) {
					log.Warn().Str("ip", ip).Int("OrgID", event.UserContext.GetOrgID()).Int("UserID", event.UserContext.GetUserID()).Msg("banned IP detected")
					connInfo.Conn.Close()
				}
			},
		}

		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		req.Header.Set("x-linkai-request-id", event.UserContext.GetTraceID())
		resp, err := c.client.Do(req)
		if err != nil {
			handleError(ctx, result, results, event, port, scheme, url, err)
			return
		}

		result.ResponseTimestamp = time.Now().UnixNano()
		body, err := findMatches(ctx, event, result, resp)
		if err != nil {
			handleError(ctx, result, results, event, port, scheme, url, err)
			return
		}

		if err := c.uploadBody(ctx, event, result, body); err != nil {
			handleError(ctx, result, results, event, port, scheme, url, err)
			return
		}

		select {
		case results <- result:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) uploadBody(ctx context.Context, event *RequestEvent, result *am.CustomWebFlowResults, body []byte) error {
	hash, link, err := c.storage.Write(ctx, event.UserContext, nil, body)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("unable to process hash/link for raw data")
		return err
	}
	result.ResponseBodyHash = hash
	result.ResponseBodyLink = link
	return nil
}

func findMatches(ctx context.Context, event *RequestEvent, result *am.CustomWebFlowResults, resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	data, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	for matchType, search := range event.Config.Match {
		switch matchType {
		case am.CustomMatchStatusCode:

			status := fmt.Sprintf("%d", resp.StatusCode)
			log.Info().Str("search", search).Str("result", status).Msg("searching for status code")
			if status == search {
				result.Result = append(result.Result, &am.CustomRequestResult{
					MatchType:    am.CustomMatchStatusCode,
					Matched:      true,
					MatchResults: []string{status},
				})
			}
		case am.CustomMatchString:
			if strings.Contains(string(data), search) {
				result.Result = append(result.Result, &am.CustomRequestResult{
					MatchType:    am.CustomMatchString,
					Matched:      true,
					MatchResults: []string{search},
				})
			}
		case am.CustomMatchRegex:
			// regexp is validated from API gateway lambda, so we should know it compiles
			regex := regexp.MustCompile(search)
			regexMatches := regex.FindAllStringSubmatch(string(data), -1)
			if len(regexMatches) > 0 {
				matches := make([]string, 0)
				for _, m := range regexMatches {
					matches = append(matches, strings.Join(m, ":"))
				}
				result.Result = append(result.Result, &am.CustomRequestResult{
					MatchType:    am.CustomMatchString,
					Matched:      true,
					MatchResults: matches,
				})
			}
		}
	}
	return data, nil
}

func handleError(ctx context.Context, result *am.CustomWebFlowResults, results chan<- *am.CustomWebFlowResults, event *RequestEvent, port int32, scheme string, url string, err error) {
	result.Error = err.Error()
	select {
	case results <- result:
		return
	case <-ctx.Done():
		return
	}
}

func buildRequest(event *RequestEvent, url string) (*http.Request, error) {

	body := strings.NewReader(event.Config.Body)
	req, err := http.NewRequest(event.Config.Method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range event.Config.Headers {
		if k == "" {
			continue
		}
		req.Header.Add(k, v)
	}

	return req, nil
}
