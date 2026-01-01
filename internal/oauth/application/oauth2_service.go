package application

import "golang.org/x/oauth2"

type OAuth2Service struct {
	config *oauth2.Config
}

func NewOAuth2Service(config *oauth2.Config) *OAuth2Service {
	return &OAuth2Service{config: config}
}
