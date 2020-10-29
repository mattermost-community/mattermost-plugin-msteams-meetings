package main

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-plugin-msteams-meetings/server/remote"

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

var oAuthMessage string = "[Click here to link your Microsoft account.](%s/plugins/" + manifest.Id + "/oauth2/connect?channelID=%s)"

func (p *Plugin) authenticateAndFetchUser(userID, userEmail, channelID string) (*msgraph.User, *authError) {
	var user *msgraph.User
	var err error

	siteURL, err := p.getSiteURL()
	if err != nil {
		p.API.LogError("authenticateAndFetchUser, cannot get site URL", "error", err.Error())
		return nil, &authError{Message: "Cannot get Site URL. Contact your sys admin.", Err: err}
	}

	userInfo, apiErr := p.store.GetUserInfo(userID)
	oauthMsg := fmt.Sprintf(
		oAuthMessage,
		siteURL, channelID)

	if apiErr != nil || userInfo == nil {
		return nil, &authError{Message: oauthMsg, Err: apiErr}
	}
	user, err = p.getUserWithToken(userInfo.OAuthToken)
	if err != nil {
		return nil, &authError{Message: oauthMsg, Err: apiErr}
	}

	return user, nil
}

func (p *Plugin) disconnect(userID string) error {
	return p.store.RemoveUser(userID)
}

func (p *Plugin) getOAuthConfig() (*oauth2.Config, error) {
	config := p.getConfiguration()

	clientID := config.OAuth2ClientID
	clientSecret := config.OAuth2ClientSecret
	clientAuthority := config.OAuth2Authority

	siteURL, err := p.getSiteURL()
	if err != nil {
		return nil, err
	}

	redirectURL := fmt.Sprintf("%s/plugins/%s/oauth2/complete", siteURL, manifest.Id)

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

	client := remote.NewClient(conf, token, p.API)

	user, err := client.GetMe()
	if err != nil {
		return nil, err
	}

	return user, nil
}
