package tokenauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/flohoss/tokenauth/pkg/cookie"
	"github.com/flohoss/tokenauth/pkg/token"
)

type Config struct {
	TokenParam    string              `json:"tokenParam,omitempty"`
	AllowedTokens []string            `json:"allowedTokens,omitempty"`
	Cookie        cookie.CookieConfig `json:"cookie,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		TokenParam: "token",
		Cookie:     cookie.DefaultCookieConfig(),
	}
}

type tokenAuth struct {
	next          http.Handler
	name          string
	tokenParam    string
	allowedTokens []string
	token         *token.Token
	cookieConfig  cookie.CookieConfig
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.AllowedTokens) == 0 {
		return nil, fmt.Errorf("allowedTokens cannot be empty")
	}

	if len(config.TokenParam) == 0 {
		return nil, fmt.Errorf("tokenParam cannot be empty")
	}

	if len(config.Cookie.Name) == 0 {
		return nil, fmt.Errorf("cookie.Name cannot be empty")
	}

	if config.Cookie.MaxAge < 0 {
		return nil, fmt.Errorf("cookie.MaxAge cannot be negative")
	}

	return &tokenAuth{
		next:          next,
		name:          name,
		tokenParam:    config.TokenParam,
		allowedTokens: config.AllowedTokens,
		token:         token.New(config.AllowedTokens),
		cookieConfig:  config.Cookie,
	}, nil
}

func (t *tokenAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	param := req.URL.Query().Get(t.tokenParam)

	if param != "" {
		if !t.token.Valid(param, false) {
			http.SetCookie(rw, cookie.Clear(t.cookieConfig))
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		authCookie := cookie.New(t.cookieConfig, token.HashToken(param))
		http.SetCookie(rw, authCookie)

		q := req.URL.Query()
		q.Del(t.tokenParam)
		req.URL.RawQuery = q.Encode()

		newURL := &url.URL{
			Scheme:   req.URL.Scheme,
			Host:     req.URL.Host,
			Path:     req.URL.Path,
			RawQuery: q.Encode(),
		}

		http.Redirect(rw, req, newURL.String(), http.StatusTemporaryRedirect)
		return
	}

	c, err := req.Cookie(t.cookieConfig.Name)
	if err == nil && t.token.Valid(c.Value, true) {
		t.next.ServeHTTP(rw, req)
		return
	}

	rw.WriteHeader(http.StatusUnauthorized)
}
