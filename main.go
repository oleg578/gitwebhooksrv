package main

import (
	"github.com/oleg578/gitwebhooksrv/config"
	logger "github.com/oleg578/loglog"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func idleTimeoutHook() func(net.Conn, http.ConnState) {
	const idleTimeout = 15 * time.Minute
	const activeTimeout = 15 * time.Minute
	var mu sync.Mutex
	m := map[net.Conn]*time.Timer{}
	return func(c net.Conn, cs http.ConnState) {
		mu.Lock()
		defer mu.Unlock()
		if t, ok := m[c]; ok {
			delete(m, c)
			t.Stop()
		}
		var d time.Duration
		switch cs {
		case http.StateNew, http.StateIdle:
			d = idleTimeout
		case http.StateActive:
			d = activeTimeout
		default:
			return
		}
		m[c] = time.AfterFunc(d, func() {
			log.Printf("closing idle conn %v after %v", c.RemoteAddr(), d)
			go func() {
				_ = c.Close()
			}()
		})
	}
}

func main() {
	//logger
	if err := logger.New(config.LogPath, "", logger.LstdFlags); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	// main tree route
	mux.Handle("/payload", http.HandlerFunc(payloadHandler))
	mux.Handle("/payload/", http.HandlerFunc(payloadHandler))
	certManager := autocert.Manager{
		Cache:      autocert.DirCache(config.CertPath),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(config.Domain),
		Email:      config.AdminMail,
	}
	srv := &http.Server{
		Addr:           ":https",
		Handler:        mux,
		ReadTimeout:    900 * time.Second,
		WriteTimeout:   900 * time.Second,
		MaxHeaderBytes: 1 << 12,
		TLSConfig:      certManager.TLSConfig(),
	}
	srv.ConnState = idleTimeoutHook()

	if err := http2.ConfigureServer(srv, &http2.Server{}); err != nil {
		logger.Fatal(err)
	}
	////https production
	go func() {
		err := http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		if err != nil {
			logger.Fatal(err)
		}
	}()
	logger.Fatal(srv.ListenAndServeTLS("", ""))
}
