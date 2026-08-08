package internal

import (
	_ "embed"
	"log/slog"
	"net/http"
)

//go:embed login.html
var loginPageHTML string

func (f *Forge) LoginHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(loginPageHTML))
}

func (f *Forge) StartLoginServer(port string) {
	addr := ":" + port
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/yunxiao/login", f.LoginHandler)
		slog.Info("login server listening", "addr", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			slog.Error("login server failed", "error", err)
		}
	}()
}
