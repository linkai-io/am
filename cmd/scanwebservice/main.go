package main

import (
	"crypto/tls"
	"flag"
	"fmt"

	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/acme/autocert"
)

const infoHTML = `<!DOCTYPE html>
<head>
    <title>Linkai's Hakken Web Asset Discovery Service Scanner</title>
    <style>
    body {
        background-color: #353b43;
        color: #abb3bb;
    }
    h2 {
        color: #e9e9e9;
    }
    a {
        color: #f3c;
    }
    div {
        font-size: 18px;
    }
    </style>
</head>
<body>
    <h2>This system is part of the Hakken Web Asset Discovery Service.</h2>
    <div>For more information please see <a href="https://linkai.io/">linkai.io</a></div>
</body>
</html>`

var (
	hostname string
	certPath string
	email    string
)

func init() {
	flag.StringVar(&hostname, "host", "scanner1.linkai.io", "hostname to use for serving files from")
	flag.StringVar(&certPath, "certs", "/opt/scanner/certs", "path to autocert cache")
	flag.StringVar(&email, "email", "service@linkai.io", "email to register with lets encrypt")
}

func main() {
	flag.Parse()

	if hostname == "" {
		log.Fatal().Msg("you did not configure a valid hostname")
	}

	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hostname),
		Cache:      autocert.DirCache(certPath),
	}

	if email != "" {
		m.Email = email
	}

	mux := http.NewServeMux()
	addRoutes(mux)

	httpsServer := &http.Server{
		Addr:         ":443",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig:    &tls.Config{GetCertificate: m.GetCertificate},
		Handler:      mux,
	}

	httpServer := &http.Server{
		Addr:         ":80",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      m.HTTPHandler(nil),
	}

	go func() {
		log.Info().Msg("Starting HTTP server")
		err := httpServer.ListenAndServe()
		if err != nil {
			log.Fatal().Msgf("error from HTTP server: %s", err)
		}
	}()

	log.Info().Msg("Starting HTTPS server")
	err := httpsServer.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatal().Msgf("error from HTTPS server: %s", err)
	}
}

func addRoutes(mux *http.ServeMux) {
	mux.Handle("/", serve())
	// add custom handlers here if necessary
}

func serve() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("remote_addr", r.RemoteAddr).Str("url", r.URL.String()).Msg("request")
		w.WriteHeader(200)
		fmt.Fprintf(w, infoHTML)
	})
}
