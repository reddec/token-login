package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/reddec/token-login/web/controllers/token"
	"github.com/reddec/token-login/web/controllers/tokens"
)

//go:embed assets/static
var static embed.FS

type Storage interface {
	token.Storage
	tokens.Storage
}

func NewAdmin(storage Storage) http.Handler {
	mux := chi.NewMux()

	staticDir, err := fs.Sub(static, "assets")
	if err != nil {
		panic(err)
	}

	mux.Mount("/", tokens.New(storage, "."))
	mux.Mount("/token/", token.New(storage, "../"))
	mux.Mount("/static/", http.FileServer(http.FS(staticDir)))
	return mux
}
