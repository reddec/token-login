package web

import (
	"net/http"
	"net/url"

	"github.com/reddec/token-login/internal/validator"
)

const (
	URLHeader           = `X-Forwarded-Uri`
	TokenHeader         = `X-Token`
	HostHeader          = "X-Forwarded-Host"
	TokenQuery          = `token`
	AuthUserHeader      = `X-User`
	AuthTokenHintHeader = `X-Token-Hint` //nolint:gosec
)

func AuthHandler(validator *validator.Validator) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestURL, err := url.Parse(request.Header.Get(URLHeader))
		if err != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		key := getToken(request, requestURL)
		host := getHost(request)
		token, err := validator.Valid(request.Context(), host, requestURL.Path, key)
		if err != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		headers := writer.Header()
		headers.Set(AuthUserHeader, token.User)
		headers.Set(AuthTokenHintHeader, token.Hint())
		for _, header := range token.Headers {
			headers.Set(header.Name, header.Value)
		}
		writer.WriteHeader(http.StatusNoContent)
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
