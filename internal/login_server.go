package internal

import (
	"log/slog"
	"net/http"
)

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
