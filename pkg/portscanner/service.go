package portscanner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// SOCKFMT is the unix domain socket to bind to
const SOCKFMT = "/opt/scanner/%s/executor.sock"

type Service struct {
	Env      string
	Sock     string
	listener net.Listener
	srv      http.Server
	scanner  *Scanner
}

func NewService() *Service {
	return &Service{scanner: New()}
}

func (s *Service) Init(env string) error {
	if err := s.scanner.Init(); err != nil {
		return err
	}

	if env == "" {
		return errors.New("environment not specificed, must be dev/prod")
	}

	s.Env = env
	s.Sock = fmt.Sprintf(SOCKFMT, s.Env)
	s.srv = http.Server{
		ReadTimeout:  time.Duration(time.Minute * 10),
		WriteTimeout: time.Duration(time.Minute * 10),
		IdleTimeout:  time.Duration(time.Minute * 30),
	}
	return nil
}

func (s *Service) Serve() error {
	var err error
	// Start Server
	os.Remove(s.Sock)

	s.listener, err = net.Listen("unix", s.Sock)

	if err := os.Chmod(s.Sock, 0770); err != nil {
		log.Fatal().Err(err).Msg("failed to change mode on socket")
	}

	if err != nil {
		log.Fatal().Err(err).Msg("Listen (UNIX socket): ")
	}

	http.HandleFunc("/scan", s.Scan)

	err = s.srv.Serve(s.listener)
	return err
}

func (s *Service) Shutdown() {
	if err := s.srv.Shutdown(context.Background()); err != nil {
		log.Info().Err(err).Msgf("HTTP server Shutdown")
	}
}

type errResponse struct {
	Error string `json:"error"`
}

func (s *Service) Scan(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		returnError(w, err.Error())
		return
	}

	request := &scanRequest{}
	err = json.Unmarshal(data, request)
	if err != nil {
		returnError(w, err.Error())
		return
	}

	timeout, cancel := context.WithTimeout(r.Context(), time.Minute*10)
	defer cancel()

	results, err := s.scanner.ScanIPv4(timeout, request.TargetIP, request.PPS, request.Ports)
	if err != nil {
		returnError(w, err.Error())
		return
	}

	data, err = json.Marshal(results)
	if err != nil {
		returnError(w, err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, string(data))
}

func returnError(w http.ResponseWriter, errMsg string) {
	w.WriteHeader(500)
	data, _ := json.Marshal(&errResponse{Error: errMsg})

	fmt.Fprintf(w, string(data))
}
