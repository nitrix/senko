package web

import (
	"context"
	"fmt"
	"net/http"
	"senko/app"
)

const DownloadDir = "downloads"

type Web struct {
	server *http.Server
}

func (w *Web) Name() string {
	return "web"
}

func (w *Web) OnRegister() {}

func (w *Web) Run(app *app.App) error {
	fs := http.FileServer(http.Dir(DownloadDir))
	http.Handle(fmt.Sprintf("/%s/", DownloadDir), http.StripPrefix(fmt.Sprintf("/%s/", DownloadDir), fs))

	w.server = &http.Server{Addr: ":80"}
	err := w.server.ListenAndServe()

	// We have to be careful, interruption of the server is returned as an error
	// but is definitely not fatal. We have other goroutines that must do proper cleanup.
	if err == http.ErrServerClosed {
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to run web server: %w", err)
	}

	return nil
}

func (w *Web) Stop() {
	_ = w.server.Shutdown(context.Background())
}

func (w *Web) SendMessage(location string, message string) error {
	return nil
}