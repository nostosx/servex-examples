package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"time"

	sdk "golang.org/x/oauth2"

	"github.com/nostosx/servex-contrib/auth/login"
	"github.com/nostosx/servex-contrib/auth/oauth2"
	"github.com/nostosx/servex-contrib/auth/oidc"
	"github.com/nostosx/servex-contrib/session"

	"github.com/nostosx/servex-examples/application/identity"
	"github.com/nostosx/servex-examples/infrastructure/config"
)

func buildLogin(ctx context.Context, cfg config.Composition, store session.Store, states oauth2.StateStore) (*login.Manager[identity.Principal], error) {
	options := []session.Option{}
	if cfg.InsecureCookies {
		options = append(options, session.WithInsecureCookies())
	}
	sessions, err := session.New(store, options...)
	if err != nil {
		return nil, err
	}
	loginConfig := login.Config[identity.Principal]{Valid: identity.Valid, SuccessURL: "/v1/notes", HTTPClient: &http.Client{Timeout: 10 * time.Second}}
	if cfg.OIDC.Issuer != "" {
		oauth, err := oauth2.New(oauth2.Config{OAuth: sdk.Config{ClientID: cfg.OIDC.ClientID, ClientSecret: cfg.OIDC.ClientSecret, RedirectURL: cfg.OIDC.RedirectURL, Scopes: []string{"openid", "profile"}, Endpoint: sdk.Endpoint{AuthURL: cfg.OIDC.AuthURL, TokenURL: cfg.OIDC.TokenURL}}, AllowInsecureLoopback: cfg.InsecureCookies}, states)
		if err != nil {
			return nil, err
		}
		// Provider key refreshes live for the application, not the startup deadline.
		var verifierOptions []oidc.Option
		issuer, _ := url.Parse(cfg.OIDC.Issuer)
		if cfg.InsecureCookies && issuer != nil && issuer.Scheme == "http" {
			host := issuer.Hostname()
			ip := net.ParseIP(host)
			if host == "localhost" || ip != nil && ip.IsLoopback() {
				verifierOptions = append(verifierOptions, oidc.WithInsecureHTTP())
			}
		}
		verifier, err := oidc.New(context.WithoutCancel(ctx), cfg.OIDC.Issuer, cfg.OIDC.ClientID, verifierOptions...)
		if err != nil {
			return nil, err
		}
		loginConfig.OIDC = &login.OIDC[identity.Principal]{Client: oauth, Verifier: verifier, Resolve: func(_ context.Context, token *oidc.IDToken) (identity.Principal, error) {
			// The issuer + subject is stable. Display claims never grant permissions.
			digest := sha256.Sum256([]byte(token.Issuer + "\x00" + token.Subject))
			var claims struct {
				Name string `json:"name"`
			}
			if err := token.Claims(&claims); err != nil {
				return identity.Principal{}, err
			}
			return identity.Principal{ID: "oidc:" + hex.EncodeToString(digest[:]), Name: claims.Name}, nil
		}}
	}
	return login.New(sessions, loginConfig)
}
