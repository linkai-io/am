package browser

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/linkai-io/am/pkg/retrier"
	"github.com/rs/zerolog/log"
	"github.com/wirepair/gcd"
)

const SOCK = "/tmp/leaser.sock"

type GcdLeaser struct {
	listener       net.Listener
	srv            http.Server
	browserLock    sync.RWMutex
	browsers       map[string]*gcd.Gcd
	browserTimeout time.Duration
}

func NewGcdLeaser() *GcdLeaser {
	return &GcdLeaser{
		browserLock:    sync.RWMutex{},
		browserTimeout: time.Second * 30,
		browsers:       make(map[string]*gcd.Gcd),
	}
}

func (g *GcdLeaser) Serve() error {
	var err error
	// Start Server
	os.Remove(SOCK)

	if err := KillOldProcesses(); err != nil {
		log.Fatal().Err(err).Msg("failed to kill old chrome processes")
	}

	if err := RemoveTmpContents(); err != nil {
		log.Fatal().Err(err).Msg("failed to remove tmp contents")
	}

	g.listener, err = net.Listen("unix", SOCK)
	// HACK HACK HACK TODO: put the users in proper 'browser' group to share file via 770
	if err := os.Chmod(SOCK, 0777); err != nil {
		log.Fatal().Err(err).Msg("failed to change mode on socket")
	}

	if err != nil {
		log.Fatal().Err(err).Msg("Listen (UNIX socket): ")
	}

	http.HandleFunc("/cleanup", g.Cleanup)
	http.HandleFunc("/acquire", g.Acquire)
	http.HandleFunc("/count", g.Count)
	http.HandleFunc("/return", g.Return)

	err = g.srv.Serve(g.listener)
	return err
}

func (g *GcdLeaser) Shutdown() {
	if err := g.srv.Shutdown(context.Background()); err != nil {
		log.Info().Err(err).Msgf("HTTP server Shutdown")
	}
}

func (g *GcdLeaser) Count(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	g.browserLock.RLock()
	count := len(g.browsers)
	g.browserLock.RUnlock()
	w.WriteHeader(200)
	fmt.Fprintf(w, strconv.Itoa(count))
}

func (g *GcdLeaser) Acquire(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	b := gcd.NewChromeDebugger()
	b.DeleteProfileOnExit()
	profileDir := randProfile()
	port := randPort()

	b.AddFlags(startupFlags)
	if err := b.StartProcess("/usr/bin/google-chrome", profileDir, port); err != nil {
		log.Error().Err(err).Msg("failed to start browser")
		w.WriteHeader(500)
		fmt.Fprintf(w, "error: "+err.Error())
		return
	}

	g.browserLock.Lock()
	g.browsers[port] = b
	g.browserLock.Unlock()
	w.WriteHeader(200)
	fmt.Fprintf(w, port)
}

func (g *GcdLeaser) Return(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	r.ParseForm()
	port := r.Form.Get("port")
	g.browserLock.Lock()
	defer g.browserLock.Unlock()

	if b, ok := g.browsers[port]; ok {
		delete(g.browsers, port)
		if err := b.ExitProcess(); err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "error: "+err.Error())
			return
		}

		w.WriteHeader(200)
		fmt.Fprintf(w, "ok")
		return
	}

	w.WriteHeader(404)
	fmt.Fprintf(w, "not found")
}

func (g *GcdLeaser) Cleanup(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if err := KillOldProcesses(); err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "failed to kill old chrome processes: "+err.Error())
		return
	}

	if err := RemoveTmpContents(); err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "failed to remove tmp contents: "+err.Error())
		return
	}
}

func RemoveTmpContents() error {
	files, err := filepath.Glob(filepath.Join("/tmp", "gcd*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func KillOldProcesses() error {
	cmd := exec.Command("killall", "google-chrome")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warn().Msgf("%s:%s", err.Error(), string(output))
	}
	return nil
}

func randPort() string {
	var l net.Listener
	retryErr := retrier.Retry(func() error {
		var err error
		l, err = net.Listen("tcp", ":0")
		return err
	})

	if retryErr != nil {
		log.Warn().Err(retryErr).Msg("unable to get port using default 9022")
		return "9022"
	}
	_, randPort, _ := net.SplitHostPort(l.Addr().String())
	l.Close()
	return randPort
}

func randProfile() string {
	profile, err := ioutil.TempDir("/tmp", "gcd")
	if err != nil {
		log.Error().Err(err).Msg("failed to create temporary profile directory")
		return "/tmp/gcd"
	}

	return profile
}
