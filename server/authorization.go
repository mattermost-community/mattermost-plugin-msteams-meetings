package main

import (
	"encoding/json"
	"fmt"
	"net/url"

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

func (p *Plugin) getOauthMessage(channelID string) (string, error) {
	pluginOauthURL, err := p.getPluginOauthURL()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("[Click here to link your Microsoft account.](%s/connect?channelID=%s)", pluginOauthURL, url.QueryEscape(channelID)), nil
}

func (p *Plugin) authenticateAndFetchUser(userID, channelID string) (*msgraph.User, *authError) {
	var user *msgraph.User
	var err error

	oauthMsg, err := p.getOauthMessage(channelID)
	if err != nil {
		p.API.LogError("authenticateAndFetchUser, cannot get oauth message", "error", err.Error())
		return nil, &authError{Message: "Error getting oauth messsage.", Err: err}
	}

	userInfo, apiErr := p.GetUserInfo(userID)
	if apiErr != nil || userInfo == nil {
		return nil, &authError{Message: oauthMsg, Err: apiErr}
	}

	user, err = p.getUserWithToken()
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

func (p *Plugin) getUserWithToken() (*msgraph.User, error) {
	user, err := p.client.GetMe()
	if err != nil {
		return nil, err
	}

	return user, nil
}
