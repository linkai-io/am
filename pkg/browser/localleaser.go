package browser

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/wirepair/gcd"
)

type LocalLeaser struct {
	browserLock    sync.RWMutex
	browsers       map[string]*gcd.Gcd
	browserTimeout time.Duration
}

func NewLocalLeaser() *LocalLeaser {
	s := &LocalLeaser{
		browserLock:    sync.RWMutex{},
		browserTimeout: time.Second * 30,
		browsers:       make(map[string]*gcd.Gcd),
	}
	return s
}

func (s *LocalLeaser) Acquire() (string, error) {
	b := gcd.NewChromeDebugger()
	b.DeleteProfileOnExit()
	profileDir := randProfile()
	port := randPort()

	b.AddFlags(startupFlags)
	if err := b.StartProcess("/usr/bin/google-chrome", profileDir, port); err != nil {
		log.Error().Err(err).Msg("failed to start browser")
		return "", err
	}
	s.browserLock.Lock()
	s.browsers[port] = b
	s.browserLock.Unlock()

	return string(port), nil
}

func (s *LocalLeaser) Return(port string) error {
	s.browserLock.Lock()
	defer s.browserLock.Unlock()

	if b, ok := s.browsers[port]; ok {
		if err := b.ExitProcess(); err != nil {
			return err
		}
		delete(s.browsers, port)
		return nil
	}

	return errors.New("not found")
}

func (s *LocalLeaser) Cleanup() (string, error) {
	if err := KillOldProcesses(); err != nil {
		return "", err
	}

	if err := RemoveTmpContents(); err != nil {
		return "", err
	}
	return "ok", nil
}
