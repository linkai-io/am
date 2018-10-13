package browser

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/linkai-io/am/am"
)

type ResponseContainer struct {
	responsesLock sync.RWMutex
	responses     map[string]*am.HTTPResponse

	readyLock sync.RWMutex
	ready     map[string]chan struct{}

	requestCount int32
}

func NewResponseContainer() *ResponseContainer {
	return &ResponseContainer{
		responses: make(map[string]*am.HTTPResponse),
		ready:     make(map[string]chan struct{}),
	}
}

// GetResponses returns all responses and clears the container
func (c *ResponseContainer) GetResponses() []*am.HTTPResponse {
	c.responsesLock.Lock()
	defer c.responsesLock.Unlock()

	r := make([]*am.HTTPResponse, len(c.responses))
	i := 0
	for _, v := range c.responses {
		r[i] = v
		i++
	}
	c.responses = make(map[string]*am.HTTPResponse, 0)
	return r
}

// Add a response to our map
func (c *ResponseContainer) Add(response *am.HTTPResponse) {
	c.responsesLock.Lock()
	c.responses[response.RequestID] = response
	c.responsesLock.Unlock()
}

func (c *ResponseContainer) IncRequest() {
	atomic.AddInt32(&c.requestCount, 1)
}

func (c *ResponseContainer) DecRequest() {
	atomic.AddInt32(&c.requestCount, -1)
}

func (c *ResponseContainer) GetRequests() int32 {
	return atomic.LoadInt32(&c.requestCount)
}

// WaitFor see if we have a readyCh for this request, if we don't, make the channel
// if we do, it is already closed so we can return
func (c *ResponseContainer) WaitFor(ctx context.Context, requestID string) error {
	var readyCh chan struct{}
	var ok bool

	defer c.remove(requestID)

	c.readyLock.Lock()
	if readyCh, ok = c.ready[requestID]; !ok {
		readyCh = make(chan struct{})
		c.ready[requestID] = readyCh
	}
	c.readyLock.Unlock()

	select {
	case <-readyCh:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// BodyReady signals WaitFor that the response is done, and we can start reading the body
func (c *ResponseContainer) BodyReady(requestID string) {
	c.readyLock.Lock()
	if _, ok := c.ready[requestID]; !ok {
		c.ready[requestID] = make(chan struct{})
	}
	close(c.ready[requestID])
	c.readyLock.Unlock()
}

// remove the request from our ready map
func (c *ResponseContainer) remove(requestID string) {
	c.readyLock.Lock()
	delete(c.ready, requestID)
	c.readyLock.Unlock()
}
