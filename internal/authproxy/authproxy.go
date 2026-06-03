// Package authproxy runs a local HTTP reverse proxy with optional HTTP Basic Auth
// in front of the app before the tunnel connects.
package authproxy

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Server is a localhost reverse proxy, optionally protected by Basic Auth.
type Server struct {
	listener net.Listener
	srv      *http.Server
	Port     int
}

// Start listens on 127.0.0.1:0 and forwards to localhost:targetPort.
// If password is non-empty, clients must send Basic Auth (user defaults to "solo").
func Start(targetPort int, user, password string) (*Server, error) {
	target, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", targetPort))
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		http.Error(w, "upstream unavailable: "+err.Error(), http.StatusBadGateway)
	}

	handler := http.Handler(proxy)
	if password != "" {
		if user == "" {
			user = "solo"
		}
		handler = basicAuth(handler, user, password)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	port := ln.Addr().(*net.TCPAddr).Port

	s := &Server{
		listener: ln,
		Port:     port,
		srv: &http.Server{
			Handler: handler,
		},
	}
	go func() { _ = s.srv.Serve(ln) }()
	return s, nil
}

// Stop shuts down the proxy server.
func (s *Server) Stop() {
	if s == nil || s.srv == nil {
		return
	}
	_ = s.srv.Shutdown(context.Background())
}

func basicAuth(next http.Handler, user, pass string) http.Handler {
	realm := "SoloEnv staging"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(u), []byte(user)) != 1 ||
			subtle.ConstantTimeCompare([]byte(p), []byte(pass)) != 1 {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
