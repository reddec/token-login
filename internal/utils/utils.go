package utils

import (
	"context"
	"encoding/base64"
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
