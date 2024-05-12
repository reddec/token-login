package utils

import (
	"context"
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
)

type userCtx struct{}

func WithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, userCtx{}, user)
}

func GetUser(ctx context.Context) string {
	v, ok := ctx.Value(userCtx{}).(string)
	if ok {
		return v
	}
	return "anonymous"
}

const flashTTL = 10

func SetFlashPath(w http.ResponseWriter, name string, value string, path string) {
	v := base64.RawStdEncoding.EncodeToString([]byte(value))
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    v,
		Path:     path,
		MaxAge:   flashTTL,
		HttpOnly: true,
	})
}

func SetFlash(w http.ResponseWriter, name string, value string) {
	SetFlashPath(w, name, value, ".")
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) string {
	http.SetCookie(w, &http.Cookie{Name: name, MaxAge: -1, HttpOnly: true, Path: "."})
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	v, err := base64.RawStdEncoding.DecodeString(c.Value)
	if err != nil {
		return ""
	}
	return string(v)
}

func Expose[T any](state func(request *http.Request) (*T, error)) *Controller[T] {
	return &Controller[T]{
		state:   state,
		actions: map[string]func(writer http.ResponseWriter, request *http.Request, state *T) error{},
	}
}

type Controller[T any] struct {
	actions map[string]func(writer http.ResponseWriter, request *http.Request, state *T) error
	state   func(request *http.Request) (*T, error)
	view    struct {
		enabled      bool
		viewTemplate *template.Template
		handler      func(writer http.ResponseWriter, request *http.Request, state *T) (any, error)
	}
}

func (ch *Controller[T]) View(viewTemplate *template.Template, handler func(writer http.ResponseWriter, request *http.Request, state *T) (any, error)) *Controller[T] {
	ch.view.viewTemplate = viewTemplate
	ch.view.handler = handler
	ch.view.enabled = true
	return ch
}

func (ch *Controller[T]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch {
	case request.Method == http.MethodGet && ch.view.enabled:
		ch.onGet(writer, request)
	case request.Method == http.MethodPost && len(ch.actions) > 0:
		ch.onPost(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (ch *Controller[T]) Action(name string, handler func(writer http.ResponseWriter, request *http.Request, state *T) error) *Controller[T] {
	ch.actions[name] = handler
	return ch
}

func (ch *Controller[T]) onGet(writer http.ResponseWriter, request *http.Request) {
	state, err := ch.state(request)
	if err != nil {
		log.Println("state:", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	view, err := ch.view.handler(writer, request, state)
	if err != nil {
		log.Println("view -", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(http.StatusOK)
	if err := ch.view.viewTemplate.Execute(writer, view); err != nil {
		log.Println("render:", err)
	}
}

func (ch *Controller[T]) onPost(writer http.ResponseWriter, request *http.Request) {
	name := request.FormValue("action")
	handler, ok := ch.actions[name]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	state, err := ch.state(request)
	if err != nil {
		log.Println("state:", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := handler(writer, request, state); err != nil {
		log.Println("exec", name, "-", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Location", ".")
	writer.WriteHeader(http.StatusSeeOther)
}
