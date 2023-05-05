package tokens

import (
	"context"
	_ "embed" // for templates
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/web/controllers/utils"
)

const (
	flashToken = "token"
)

//go:embed index.gohtml
var source string

type Storage interface {
	ListTokens(ctx context.Context, user string) ([]*dbo.Token, error)
	CreateToken(ctx context.Context, params dbo.TokenParams) error
	DeleteToken(ctx context.Context, ref dbo.TokenRef) error
	UpdateTokenKey(ctx context.Context, ref dbo.TokenRef, key dbo.Key) error
}

func New(store Storage, rootPath string) http.Handler {
	srv := utils.Expose[State](func(request *http.Request) (*State, error) {
		return &State{
			User: utils.GetUser(request),
		}, nil
	})

	// render main page
	srv.View(template.Must(template.New("").Funcs(map[string]any{
		"globals": func() map[string]any {
			return map[string]any{
				"rootPath": rootPath,
			}
		},
	}).Parse(source)), func(writer http.ResponseWriter, request *http.Request, state *State) (any, error) {
		tokens, err := store.ListTokens(request.Context(), state.User)
		if err != nil {
			return nil, fmt.Errorf("list token: %w", err)
		}

		newToken := utils.GetFlash(writer, request, flashToken)

		return &viewContext{
			Tokens: tokens,
			Token:  newToken,
			State:  state,
		}, nil
	})

	// create token
	srv.Action("", func(writer http.ResponseWriter, request *http.Request, state *State) error {
		key, err := dbo.NewKey()

		if err != nil {
			return fmt.Errorf("generate key: %w", err)
		}

		if err := store.CreateToken(request.Context(), dbo.TokenParams{
			User: state.User,
			Config: dbo.TokenConfig{
				Label: request.FormValue("label"),
				Path:  request.FormValue("path"),
			},
			Key: key,
		}); err != nil {
			return fmt.Errorf("save key: %w", err)
		}

		utils.SetFlash(writer, flashToken, key.String())
		return nil
	})

	// delete token
	srv.Action("delete", func(writer http.ResponseWriter, request *http.Request, state *State) error {
		id, err := strconv.ParseInt(request.FormValue("token"), 10, 64)

		if err != nil {
			return fmt.Errorf("parse ID: %w", err)
		}

		if err := store.DeleteToken(request.Context(), dbo.TokenRef{User: state.User, ID: id}); err != nil {
			return fmt.Errorf("delete token: %w", err)
		}

		return nil
	})

	// refresh token
	srv.Action("refresh", func(writer http.ResponseWriter, request *http.Request, state *State) error {
		id, err := strconv.ParseInt(request.FormValue("token"), 10, 64)

		if err != nil {
			return fmt.Errorf("parse ID: %w", err)
		}

		key, err := dbo.NewKey()
		if err != nil {
			return fmt.Errorf("generate key: %w", err)
		}

		if err := store.UpdateTokenKey(request.Context(), dbo.TokenRef{User: state.User, ID: id}, key); err != nil {
			return fmt.Errorf("get token: %w", err)
		}

		utils.SetFlash(writer, flashToken, key.String())
		return nil
	})

	return srv
}

type State struct {
	User string
}

type viewContext struct {
	Token  string
	Tokens []*dbo.Token
	*State
}

func (vc *viewContext) Hint() string {
	if len(vc.Token) >= dbo.HintChars {
		return vc.Token[:dbo.HintChars]
	}
	return ""
}

func (vc *viewContext) Payload() string {
	if len(vc.Token) < dbo.HintChars {
		return vc.Token
	}
	return vc.Token[dbo.HintChars:]
}

func (vc *viewContext) CreatedToken() *dbo.Token {
	h := vc.Hint()
	for _, v := range vc.Tokens {
		if v.Hint() == h {
			return v
		}
	}
	return nil
}
