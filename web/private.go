package web

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/reddec/token-login/internal/cache"
	"github.com/reddec/token-login/internal/types"
)

const (
	URLHeader           = `X-Forwarded-Uri`
	TokenHeader         = `X-Token`
	HostHeader          = "X-Forwarded-Host"
	TokenQuery          = `token`
	AuthUserHeader      = `X-User`
	AuthTokenHintHeader = `X-Token-Hint` //nolint:gosec
)

type Hit struct {
	Time time.Time
	ID   int
}

func AuthHandler(state *cache.Cache, accessLog chan<- Hit) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestURL, err := url.Parse(request.Header.Get(URLHeader))
		if err != nil {
			slog.Debug("failed parse request url", "error", err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		rawKey := getToken(request, requestURL)
		host := getHost(request)
		key, err := types.ParseKey(rawKey)
		if err != nil {
			slog.Debug("failed parse key", "error", err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, found := state.FindByKey(key.ID())
		if !found {
			slog.Debug("token not found", "key", key.ID())
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		if ok := token.AccessKey.Valid(host, requestURL.Path, key.Payload()); !ok {
			slog.Debug("access key invalid", "key", key.ID())
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		headers := writer.Header()
		headers.Set(AuthUserHeader, token.DBToken.User)
		headers.Set(AuthTokenHintHeader, key.ID().String())
		for _, header := range token.DBToken.Headers {
			headers.Set(header.Name, header.Value)
		}
		writer.WriteHeader(http.StatusNoContent)
		select {
		case accessLog <- Hit{Time: time.Now(), ID: token.DBToken.ID}:
		default:
		}
	})
}

func getToken(req *http.Request, sourceURL *url.URL) string {
	if apiKey := req.Header.Get(TokenHeader); apiKey != "" {
		return apiKey
	}
	if apiKey := sourceURL.Query().Get(TokenQuery); apiKey != "" {
		return apiKey
	}
	return ""
}

func getHost(req *http.Request) string {
	if host := req.Header.Get(HostHeader); host != "" {
		return host
	}
	return req.Host
}
