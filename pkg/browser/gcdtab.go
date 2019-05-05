package browser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/parsers"
	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
)

var (
	ErrNavigationTimedOut = errors.New("navigation timed out")
	ErrTabCrashed         = errors.New("tab crashed")
	ErrTabClosing         = errors.New("closing")
	ErrTimedOut           = errors.New("request timed out")
	ErrNavigating         = errors.New("error in navigation")
)

type Tab struct {
	t                     *gcd.ChromeTarget
	address               *am.ScanGroupAddress
	container             *ResponseContainer
	crashedCh             chan string
	exitCh                chan struct{}
	navigationCh          chan int
	lastNodeChangeTimeVal atomic.Value
}

func NewTab(ctx context.Context, tab *gcd.ChromeTarget, address *am.ScanGroupAddress) *Tab {
	t := &Tab{
		t:            tab,
		address:      address,
		container:    NewResponseContainer(),
		crashedCh:    make(chan string),
		exitCh:       make(chan struct{}),
		navigationCh: make(chan int),
	}
	t.subscribeBrowserEvents(ctx)
	return t
}

// Close the exit channel
func (t *Tab) Close() {
	close(t.exitCh)
}

// LoadPage capture network traffic and take screen shot of DOM and image
func (t *Tab) LoadPage(ctx context.Context, url string) error {
	navParams := &gcdapi.PageNavigateParams{Url: url, TransitionType: "typed"}
	log.Ctx(ctx).Info().Msg("navigating")
	_, _, errText, err := t.t.Page.NavigateWithParams(navParams)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("host_address", t.address.HostAddress).
			Str("ip_address", t.address.IPAddress).
			Str("url", url).Msg("failed to load page")
		return err
	}

	if errText != "" {
		return errors.Wrap(ErrNavigating, errText)
	}

	log.Ctx(ctx).Info().Str("url", url).Str("err_text", errText).Msg("navigating complete")
	err = t.WaitReady(ctx, time.Second*3)
	log.Ctx(ctx).Info().Msg("wait ready returned")
	return err
}

// InjectJS only caller knows what the response type will be so return an interface{}
// caller must type check to whatever they expect
func (t *Tab) InjectJS(inject string) (interface{}, error) {
	params := &gcdapi.RuntimeEvaluateParams{
		Expression:            inject,
		ObjectGroup:           "linkai",
		IncludeCommandLineAPI: false,
		Silent:                true,
		ReturnByValue:         true,
		GeneratePreview:       false,
		UserGesture:           false,
		AwaitPromise:          false,
		ThrowOnSideEffect:     false,
		Timeout:               1000,
	}
	r, exp, err := t.t.Runtime.EvaluateWithParams(params)
	if err != nil {
		return nil, err
	}
	if exp != nil {
		log.Warn().Err(err).Msg("failed to inject script")
	}

	return r.Value, nil
}

// GetURL by looking at the navigation history
func (t *Tab) GetURL(ctx context.Context) string {
	_, entries, err := t.t.Page.GetNavigationHistory()
	if err != nil || len(entries) == 0 {
		return ""
	}
	return entries[len(entries)-1].Url
}

// WaitReady waits for the page to load, DOM to be stable, and no network traffic in progress
func (t *Tab) WaitReady(ctx context.Context, stableAfter time.Duration) error {
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	navTimer := time.After(30 * time.Second)
	log.Ctx(ctx).Info().Msg("waiting for nav to complete")
	// wait navigation to complete.
	select {
	case <-navTimer:
		return ErrNavigationTimedOut
	case <-ctx.Done():
		return ctx.Err()
	case <-t.exitCh:
		return errors.New("exiting")
	case reason := <-t.crashedCh:
		return errors.Wrap(ErrTabCrashed, reason)
	case <-t.navigationCh:
	}

	stableTimer := time.After(5 * time.Second)

	// wait for DOM & network stability
	log.Ctx(ctx).Info().Msg("waiting for DOM & network stability")
	for {
		select {
		case reason := <-t.crashedCh:
			return errors.Wrap(ErrTabCrashed, reason)
		case <-ctx.Done():
			return ctx.Err()
		case <-t.exitCh:
			return ErrTabClosing
		case <-stableTimer:
			log.Ctx(ctx).Info().Msg("stability timed out")
			return ErrTimedOut
		case <-ticker.C:
			if changeTime, ok := t.lastNodeChangeTimeVal.Load().(time.Time); ok {
				//log.Info().Int32("requests", t.container.GetRequests()).Msgf("tick %s", time.Now().Sub(changeTime))
				if time.Now().Sub(changeTime) >= stableAfter && t.container.GetRequests() == 0 {
					// times up, should be stable now
					log.Ctx(ctx).Info().Msg("stable")
					return nil
				}
			}
		}
	}
}

// TakeScreenshot returns a png image, base64 encoded, or error if failed
func (t *Tab) TakeScreenshot(ctx context.Context) (string, error) {
	params := &gcdapi.PageCaptureScreenshotParams{
		Format:  "png",
		Quality: 100,
		Clip: &gcdapi.PageViewport{
			X:      0,
			Y:      0,
			Width:  1024,
			Height: 768,
			Scale:  float64(1)},
		FromSurface: true,
	}

	return t.t.Page.CaptureScreenshotWithParams(params)
}

// SerializeDOM and return it as string
func (t *Tab) SerializeDOM() string {
	node, err := t.t.DOM.GetDocument(-1, true)
	if err != nil {
		return ""
	}
	html, err := t.t.DOM.GetOuterHTMLWithParams(&gcdapi.DOMGetOuterHTMLParams{
		NodeId: node.NodeId,
	})
	if err != nil {
		return ""
	}
	return html
}

// GetNetworkTraffic returns all responses after page load
func (t *Tab) GetNetworkTraffic() (*am.HTTPResponse, []*am.HTTPResponse) {
	return t.container.GetResponses()
}

// CaptureNetworkTraffic ensures we capture all traffic (only saving text bodies) during navigation.
func (t *Tab) CaptureNetworkTraffic(ctx context.Context, address *am.ScanGroupAddress, port string) {

	t.t.Network.EnableWithParams(&gcdapi.NetworkEnableParams{
		MaxPostDataSize:       -1,
		MaxResourceBufferSize: -1,
		MaxTotalBufferSize:    -1,
	})

	// sanity check
	if _, err := t.t.Network.SetBlockedURLs(parsers.BannedURLs); err != nil {
		log.Warn().Err(err).Msg("unable to set banned urls")
	}

	t.t.Subscribe("network.loadingFailed", func(target *gcd.ChromeTarget, payload []byte) {
		log.Info().Msgf("failed: %s\n", string(payload))
	})

	t.t.Subscribe("Network.requestWillBeSent", func(target *gcd.ChromeTarget, payload []byte) {
		message := &gcdapi.NetworkRequestWillBeSentEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			return
		}
		//message.Params.RedirectResponse.RemoteIPAddress
		if message.Params.Type == "Document" {
			t.container.SetLoadRequest(message.Params.RequestId)
		}
	})

	t.t.Subscribe("Network.responseReceived", func(target *gcd.ChromeTarget, payload []byte) {
		//log.Info().Msgf("RESPONSE DATA: %#v\n", string(payload))
		defer t.container.DecRequest()
		t.container.IncRequest()

		message := &gcdapi.NetworkResponseReceivedEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			return
		}

		if parsers.IsBannedIP(message.Params.Response.RemoteIPAddress) {
			log.Ctx(ctx).Warn().Str("url", message.Params.Response.Url).Str("ip_address", message.Params.Response.RemoteIPAddress).Msg("BANNED IP REQUESTED")
			return
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		p := message.Params
		url := p.Response.Url

		if strings.HasPrefix(p.Response.Url, "data") {
			url = "(dataurl)"
		}

		log.Ctx(ctx).Info().Str("request_id", p.RequestId).Str("url", url).Msg("waiting")
		if err := t.container.WaitFor(timeoutCtx, p.RequestId); err != nil {
			return
		}

		// ignore data urls
		if strings.HasPrefix(p.Response.Url, "data") {
			return
		}

		response := t.buildResponse(address, port, message)
		t.container.Add(response)
	})

	t.t.Subscribe("Network.loadingFinished", func(target *gcd.ChromeTarget, payload []byte) {
		//log.Info().Msgf("loadingFinished DATA: %#v\n", string(payload))
		message := &gcdapi.NetworkLoadingFinishedEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			return
		}
		log.Ctx(ctx).Info().Str("request_id", message.Params.RequestId).Msg("finished")
		t.container.BodyReady(message.Params.RequestId)
	})
}

// buildResponse fills out a new am.HTTPResponse with all relevant details
func (t *Tab) buildResponse(address *am.ScanGroupAddress, requestedPort string, message *gcdapi.NetworkResponseReceivedEvent) *am.HTTPResponse {
	var host string
	var responsePort string
	var scheme string

	p := message.Params

	if parsers.IsBannedIP(p.Response.RemoteIPAddress) {
		return nil
	}

	u, err := url.Parse(p.Response.Url)
	if err != nil {
		log.Warn().
			Err(err).
			Str("host_address", address.HostAddress).
			Str("ip_address", address.IPAddress).
			Str("port", requestedPort).
			Msg("failed to parse url, results may be inaccurate")

		host = ""
		responsePort = requestedPort
		scheme = "http"
	} else {
		host = u.Host
		scheme = u.Scheme
		responsePort = u.Port()
		if responsePort == "" {
			if u.Scheme == "http" {
				responsePort = "80"
			} else if u.Scheme == "https" {
				responsePort = "443"
			}
		}
	}

	response := &am.HTTPResponse{
		Scheme:            scheme,
		AddressHash:       convert.HashAddress(p.Response.RemoteIPAddress, host),
		IPAddress:         p.Response.RemoteIPAddress,
		HostAddress:       host,
		RequestedPort:     requestedPort,
		ResponsePort:      responsePort,
		RequestID:         p.RequestId,
		URL:               p.Response.Url,
		Headers:           t.encodeHeaders(p.Response.Headers),
		MimeType:          p.Response.MimeType,
		Status:            p.Response.Status,
		StatusText:        p.Response.StatusText,
		RawBody:           t.encodeResponseBody(message),
		WebCertificate:    t.extractCertificate(p.Response.RemoteIPAddress, host, responsePort, message),
		ResponseTimestamp: time.Now().UnixNano(),
	}

	// set additional properties of web certificate

	if p.Type == "Document" {
		response.IsDocument = true
	}

	return response
}

// encode the header depending on type, and lower case the header name so easier to search in DB.
func (t *Tab) encodeHeaders(gcdHeaders map[string]interface{}) map[string]string {
	headers := make(map[string]string, len(gcdHeaders))
	for k, v := range gcdHeaders {
		name := strings.ToLower(k)
		switch rv := v.(type) {
		case string:
			headers[name] = rv
		case []string:
			headers[name] = strings.Join(rv, ",")
		case nil:
			headers[name] = ""
		default:
			log.Warn().Str("header_name", k).Msg("unable to encode header value")
		}
	}
	return headers
}

func (t *Tab) extractCertificate(ipAddress, host, port string, message *gcdapi.NetworkResponseReceivedEvent) *am.WebCertificate {
	p := message.Params

	u, err := url.Parse(p.Response.Url)
	if err != nil {
		return nil
	}

	if u.Hostname() == t.address.HostAddress && u.Scheme == "https" &&
		strings.HasPrefix(p.Response.Url, "https") && p.Response.SecurityDetails != nil {

		cert := convert.NetworkCertificateToWebCertificate(p.Response.SecurityDetails)
		cert.AddressHash = convert.HashAddress(ipAddress, host)
		cert.IPAddress = ipAddress
		cert.Port = port
		return cert
	}

	return nil
}

func (t *Tab) encodeResponseBody(p *gcdapi.NetworkResponseReceivedEvent) string {

	var err error
	var encoded bool
	var body []byte
	var bodyStr string

	bodyStr, encoded, err = t.t.Network.GetResponseBody(p.Params.RequestId)
	if err != nil {
		log.Warn().Str("url", p.Params.Response.Url).Err(err).Msg("failed to get body")
	}

	body = []byte(bodyStr)
	if encoded {
		body, _ = base64.StdEncoding.DecodeString(bodyStr)
	}

	// we don't want to capture anything other than text based files.
	if !strings.HasPrefix(http.DetectContentType(body), "text") {
		bodyStr = ""
	}

	return bodyStr
}

func (t *Tab) domUpdated(ctx context.Context) func(target *gcd.ChromeTarget, payload []byte) {
	return func(target *gcd.ChromeTarget, payload []byte) {
		log.Ctx(ctx).Info().Msg("dom updated")
		t.lastNodeChangeTimeVal.Store(time.Now())
	}
}

func (t *Tab) subscribeBrowserEvents(ctx context.Context) {
	t.t.DOM.Enable()
	t.t.Inspector.Enable()
	t.t.Page.Enable()
	t.t.Security.Enable()

	t.t.Security.SetOverrideCertificateErrors(true)

	t.t.Subscribe("Security.certificateError", func(target *gcd.ChromeTarget, payload []byte) {
		resp := &gcdapi.SecurityCertificateErrorEvent{}
		err := json.Unmarshal(payload, resp)
		if err != nil {
			return
		}
		log.Info().Str("type", resp.Params.ErrorType).Msg("handling certificate error")
		p := &gcdapi.SecurityHandleCertificateErrorParams{
			EventId: resp.Params.EventId,
			Action:  "continue",
		}

		t.t.Security.HandleCertificateErrorWithParams(p)
		log.Ctx(ctx).Info().Msg("certificate error handled")
	})

	t.t.Subscribe("Inspector.targetCrashed", func(target *gcd.ChromeTarget, payload []byte) {
		log.Ctx(ctx).Warn().Msgf("tab crashed: %s", string(payload))
		select {
		case t.crashedCh <- "crashed":
		case <-t.exitCh:
		}
	})

	t.t.Subscribe("Inspector.detached", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.InspectorDetachedEvent{}
		err := json.Unmarshal(payload, header)
		reason := "detached"

		if err == nil {
			reason = header.Params.Reason
		}

		select {
		case t.crashedCh <- reason:
		case <-t.exitCh:
		}
	})

	t.t.Subscribe("Page.loadEventFired", func(target *gcd.ChromeTarget, payload []byte) {
		select {
		case t.navigationCh <- 0:
		case <-t.exitCh:
		}
	})

	// new nodes
	t.t.Subscribe("DOM.setChildNodes", t.domUpdated(ctx))
	t.t.Subscribe("DOM.attributeModified", t.domUpdated(ctx))
	t.t.Subscribe("DOM.attributeRemoved", t.domUpdated(ctx))
	t.t.Subscribe("DOM.characterDataModified", t.domUpdated(ctx))
	t.t.Subscribe("DOM.childNodeCountUpdated", t.domUpdated(ctx))
	t.t.Subscribe("DOM.childNodeInserted", t.domUpdated(ctx))
	t.t.Subscribe("DOM.childNodeRemoved", t.domUpdated(ctx))
	t.t.Subscribe("DOM.documentUpdated", t.domUpdated(ctx))

}
