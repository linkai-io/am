package certstream

import (
	"strings"
	"sync"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/parsers"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type CertListener struct {
	etlds    map[string]struct{}
	etldLock *sync.RWMutex
	closeCh  chan struct{}
	batcher  *Batcher
}

func New(batcher *Batcher) *CertListener {
	l := &CertListener{}
	l.etlds = make(map[string]struct{})
	l.etldLock = &sync.RWMutex{}
	l.batcher = batcher
	return l
}

func (l *CertListener) Init(closeCh chan struct{}) error {
	l.closeCh = closeCh
	go l.runListener()
	return nil
}

func (l *CertListener) AddETLD(etld string) {
	l.etldLock.Lock()
	l.etlds[strings.ToLower(etld)] = struct{}{}
	l.etldLock.Unlock()
}

func (l *CertListener) HasETLD(domain string) (string, bool) {
	etld, err := parsers.GetETLD(domain)
	if err != nil {
		return "", false
	}

	l.etldLock.Lock()
	_, exist := l.etlds[etld]
	l.etldLock.Unlock()

	return etld, exist
}

func (l *CertListener) runListener() {
	// The false flag specifies that we want heartbeat messages.
	stream, errStream := CertStreamEventStream(l.closeCh)

	for {
		select {
		case entry := <-stream:
			if entry.MessageType != "certificate_update" {
				continue
			}

			if entry.Data.LeafCert.Subject.CN != "" {
				l.formatAdd(entry.Data.LeafCert.Subject.CN)
			}

			if len(entry.Data.LeafCert.AllDomains) > 0 {
				for _, subdomain := range entry.Data.LeafCert.AllDomains {
					l.formatAdd(subdomain)
				}
			}
		case err := <-errStream:
			log.Error().Err(err).Msg("error")
		case <-l.closeCh:
			return
		}
	}
}

func (l *CertListener) formatAdd(subdomain string) {
	subdomain = strings.TrimLeft(strings.ToLower(subdomain), "*.")
	//log.Debug().Str("subdomain", subdomain).Msg("new host")
	if etld, exist := l.HasETLD(subdomain); exist {
		log.Info().Str("subdomain", subdomain).Msg("new host")
		l.batcher.Add(&am.CTSubdomain{ETLD: etld, Subdomain: subdomain, InsertedTime: time.Now().UnixNano()})
	}
}

func CertStreamEventStream(closeCh chan struct{}) (chan *CertStreamEntry, chan error) {
	outputStream := make(chan *CertStreamEntry)
	errStream := make(chan error)
	go func(closeCh chan struct{}, outputStream chan *CertStreamEntry, errStream chan error) {
		for {
			runLoop(closeCh, outputStream, errStream)
		}
	}(closeCh, outputStream, errStream)

	return outputStream, errStream
}

func runLoop(closeCh chan struct{}, outputStream chan *CertStreamEntry, errStream chan error) {
	c, _, err := websocket.DefaultDialer.Dial("wss://certstream.calidog.io", nil)
	if err != nil {
		errStream <- errors.Wrap(err, "Error connecting to certstream! Sleeping a few seconds and reconnecting... ")
		time.Sleep(5 * time.Second)
		return
	}

	for {
		var v *CertStreamEntry
		err := c.ReadJSON(&v)
		if err != nil {
			errStream <- errors.Wrap(err, "Error decoding json frame!")
			c.Close()
			c = nil
			break
		}

		if v.MessageType == "" {
			errStream <- errors.New("could not create cert object. Malformed json input recieved. Skipping")
			continue
		}

		if v.MessageType == "heartbeat" {
			continue
		}
		select {
		case <-closeCh:
			return
		case outputStream <- v:
		}
	}
}
