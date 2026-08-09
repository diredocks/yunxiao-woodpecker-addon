// Package internal implements the HTTP proxy server for login page rendering
// and webhook forwarding for the woodpecker forge addon.
package internal

import (
	_ "embed"
	"io"
	"log/slog"
	"maps"
	"net/http"
)

//go:embed login.html
var loginPageHTML string

func (f *Forge) LoginHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(loginPageHTML))
}

func (f *Forge) HookProxyHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "missing target", http.StatusBadRequest)
		return
	}

	r.Header.Del("X-Gitlab-Token")

	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
	if err != nil {
		slog.Error("hook proxy: failed to create request", "error", err, "target", target)
		http.Error(w, "proxy error", http.StatusInternalServerError)
		return
	}

	proxyReq.Header = r.Header

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		slog.Error("hook proxy: failed to forward", "error", err, "target", target)
		http.Error(w, "proxy error", http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	maps.Copy(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		slog.Error("hook proxy: failed to copy response", "error", err)
	}
}

func (f *Forge) StartProxyServer(port string) {
	addr := ":" + port
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/yunxiao/login", f.LoginHandler)
		mux.HandleFunc("/yunxiao/hook", f.HookProxyHandler)
		slog.Info("proxy server listening", "addr", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			slog.Error("proxy server failed", "error", err)
		}
	}()
}
