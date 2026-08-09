// Package internal implements the HTTP proxy server for login page rendering
// and webhook forwarding for the woodpecker forge addon.
package internal

import (
	_ "embed"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

//go:embed login.html
var loginPageHTML string

func (f *Forge) LoginHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(loginPageHTML))
}

func (f *Forge) HookProxyHandler() http.Handler {
	backend, _ := url.Parse(f.WoodpeckerHost)
	if backend.Scheme == "" {
		backend.Scheme = "http"
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Header.Del("X-Gitlab-Token")
			req.URL.Scheme = backend.Scheme
			req.URL.Host = backend.Host
			req.URL.Path = "/api/hook"
			req.Host = backend.Host
		},
	}
}

func (f *Forge) StartProxyServer(port string) {
	addr := ":" + port
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/yunxiao/login", f.LoginHandler)
		mux.Handle("/yunxiao/hook", f.HookProxyHandler())
		slog.Info("proxy server listening", "addr", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			slog.Error("proxy server failed", "error", err)
		}
	}()
}
