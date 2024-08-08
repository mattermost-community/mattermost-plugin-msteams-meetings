package main

import (
	"encoding/json"
	"fmt"

	msgraph "github.com/yaegashi/msgraph.go/beta"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

type authError struct {
	Message string `json:"message"`
	Err     error  `json:"err"`
}

func (ae *authError) Error() string {
	errorString, _ := json.Marshal(ae)
	return string(errorString)
}

func (p *Plugin) getOauthMessage() (string, error) {
	pluginOauthURL, err := p.getPluginOauthURL()
	if err != nil {
		return "", err
	}

	return "[Click here to link your Microsoft account.](" + pluginOauthURL + "/connect?channelID=%s)", nil
}

func (p *Plugin) authenticateAndFetchUser(userID, channelID string) (*msgraph.User, *authError) {
	var user *msgraph.User
	var err error

	oAuthMessage, err := p.getOauthMessage()
	if err != nil {
		p.API.LogError("authenticateAndFetchUser, cannot get oauth message", "error", err.Error())
		return nil, &authError{Message: "Error getting oauth messsage.", Err: err}
	}

	oauthMsg := fmt.Sprintf(oAuthMessage, channelID)

	userInfo, apiErr := p.GetUserInfo(userID)
	if apiErr != nil || userInfo == nil {
		return nil, &authError{Message: oauthMsg, Err: apiErr}
	}
	user, err = p.getUserWithToken(userInfo.OAuthToken)
	if err != nil {
		return nil, &authError{Message: oauthMsg, Err: err}
	}

	return user, nil
}

func (p *Plugin) disconnect(userID string) error {
	return p.RemoveUser(userID)
}

func (p *Plugin) getOAuthConfig() (*oauth2.Config, error) {
	config := p.getConfiguration()

	clientID := config.OAuth2ClientID
	clientSecret := config.OAuth2ClientSecret
	clientAuthority := config.OAuth2Authority

	pluginOauthURL, err := p.getPluginOauthURL()
	if err != nil {
		return nil, err
	}

	redirectURL := fmt.Sprintf("%s/complete", pluginOauthURL)

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"offline_access",
			"OnlineMeetings.ReadWrite",
		},
		Endpoint: microsoft.AzureADEndpoint(clientAuthority),
	}, nil
}

func (p *Plugin) getUserWithToken(token *oauth2.Token) (*msgraph.User, error) {
	conf, err := p.getOAuthConfig()
	if err != nil {
		return nil, err
	}

	client := p.NewClient(conf, token)

	user, err := client.GetMe()
	if err != nil {
		return nil, err
	}

	return user, nil
}
