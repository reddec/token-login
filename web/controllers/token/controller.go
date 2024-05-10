package token

import (
	"context"
	_ "embed" // for templates
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/web/controllers/utils"
)

//go:embed index.gohtml
var source string

type State struct {
	User  string
	Token *dbo.Token
	Ref   dbo.TokenRef
}

type Storage interface {
	GetToken(ctx context.Context, ref dbo.TokenRef) (*dbo.Token, error)
	UpdateTokenConfig(ctx context.Context, ref dbo.TokenRef, config dbo.TokenConfig) error
}

func New(store Storage, rootPath string) http.Handler {
	srv := utils.Expose[State](func(request *http.Request) (*State, error) {
		id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse ID: %w", err)
		}
		user := utils.GetUser(request)
		ref := dbo.TokenRef{User: user, ID: id}
		token, err := store.GetToken(request.Context(), ref)
		if err != nil {
			return nil, fmt.Errorf("find token: %w", err)
		}
		return &State{
			User:  user,
			Token: token,
			Ref:   ref,
		}, nil
	})

	// render main page
	srv.View(template.Must(template.New("").Funcs(map[string]any{
		"globals": func() map[string]any {
			return map[string]any{
				"rootPath": path.Join(rootPath, ".."),
			}
		},
	}).Parse(source)), func(_ http.ResponseWriter, _ *http.Request, state *State) (any, error) {
		return &tokenHeadersContext{
			State: state,
		}, nil
	})

	// update token
	srv.Action("", func(_ http.ResponseWriter, request *http.Request, state *State) error {
		label := strings.TrimSpace(request.FormValue("label"))
		allowedPath := strings.TrimSpace(request.FormValue("path"))
		host := strings.TrimSpace(request.FormValue("host"))
		return store.UpdateTokenConfig(request.Context(), state.Ref, dbo.TokenConfig{
			Label:   label,
			Path:    allowedPath,
			Host:    host,
			Headers: state.Token.Headers,
		})
	})

	// add header
	srv.Action("headers", func(_ http.ResponseWriter, request *http.Request, state *State) error {
		name := strings.TrimSpace(request.FormValue("name"))
		value := strings.TrimSpace(request.FormValue("value"))

		return store.UpdateTokenConfig(request.Context(), state.Ref, dbo.TokenConfig{
			Label:   state.Token.Label,
			Path:    state.Token.Path,
			Headers: state.Token.Headers.Without(name).With(name, value),
		})
	})

	// remove header
	srv.Action("headersDelete", func(_ http.ResponseWriter, request *http.Request, state *State) error {
		name := strings.TrimSpace(request.FormValue("name"))

		return store.UpdateTokenConfig(request.Context(), state.Ref, dbo.TokenConfig{
			Label:   state.Token.Label,
			Path:    state.Token.Path,
			Headers: state.Token.Headers.Without(name),
		})
	})

	mux := chi.NewMux()
	mux.Mount("/{id}/", srv)
	return mux
}

type tokenHeadersContext struct {
	*State
}
