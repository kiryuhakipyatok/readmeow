package oauth

import (
	"readmeow/internal/config"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type OAuthConfig struct {
	GoogleOAuthConfig *oauth2.Config
	GithubOAuthConfig *oauth2.Config
	StateExpire       time.Duration
}

type oauthParams struct {
	ClientId     string
	ClientSecret string
	RedirectURL  string
}

func newGoogleOAuth(op oauthParams) *oauth2.Config {
	googleOAuthConf := &oauth2.Config{
		ClientID:     op.ClientId,
		ClientSecret: op.ClientSecret,
		Scopes:       []string{"email", "profile"},
		RedirectURL:  op.RedirectURL,
		Endpoint:     google.Endpoint,
	}
	return googleOAuthConf
}

func newGithubOAuth(op oauthParams) *oauth2.Config {
	githubOAuthConf := &oauth2.Config{
		ClientID:     op.ClientId,
		ClientSecret: op.ClientSecret,
		Scopes:       []string{"user:email"},
		RedirectURL:  op.RedirectURL,
		Endpoint:     github.Endpoint,
	}
	return githubOAuthConf
}

func NewOAuthConfig(cfg config.OAuthConfig) OAuthConfig {
	googleOAuthParams := oauthParams{
		ClientId:     cfg.GoogleClientId,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
	}
	githubOAuthParams := oauthParams{
		ClientId:     cfg.GithubClientId,
		ClientSecret: cfg.GithubClientSecret,
		RedirectURL:  cfg.GithubRedirectURL,
	}
	githubOAuthConfig := newGithubOAuth(githubOAuthParams)
	googleOAuthConfig := newGoogleOAuth(googleOAuthParams)
	return OAuthConfig{
		GoogleOAuthConfig: googleOAuthConfig,
		GithubOAuthConfig: githubOAuthConfig,
		StateExpire:       cfg.StateTTL,
	}
}
