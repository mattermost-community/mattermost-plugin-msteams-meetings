// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/mattermost/mattermost-plugin-mstelephony/server/remote"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	msgraph "github.com/yaegashi/msgraph.go/beta"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

const (
	postMeetingKey = "post_meeting_"

	botUserName    = "mstmeetings"
	botDisplayName = "MS Teams Meetings"
	botDescription = "Created by the MS Teams Meetings plugin."

	tokenKey           = "token_"
	tokenKeyByRemoteID = "tbyrid_"

	stateLength = 3
)

var oAuthMessage string = "[Click here to link your Microsoft account.](%s/plugins/" + manifest.Id + "/oauth2/connect?channelID=%s)"

type Plugin struct {
	plugin.MattermostPlugin

	// botUserID of the created bot account.
	botUserID string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// OnActivate checks if the configurations is valid and ensures the bot account exists
func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		return err
	}

	if _, err := p.getSiteUrl(); err != nil {
		return err
	}

	botUserID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    botUserName,
		DisplayName: botDisplayName,
		Description: botDescription,
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}
	p.botUserID = botUserID

	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "couldn't get bundle path")
	}

	if err = p.API.RegisterCommand(getCommand()); err != nil {
		return errors.WithMessage(err, "OnActivate: failed to register command")
	}

	profileImage, err := ioutil.ReadFile(filepath.Join(bundlePath, "assets", "profile.png"))
	if err != nil {
		return errors.Wrap(err, "couldn't read profile image")
	}

	if appErr := p.API.SetProfileImage(botUserID, profileImage); appErr != nil {
		return errors.Wrap(appErr, "couldn't set profile image")
	}

	return nil
}

func (p *Plugin) getSiteUrl() (string, error) {
	siteUrlRef := p.API.GetConfig().ServiceSettings.SiteURL
	if siteUrlRef == nil || *siteUrlRef == "" {
		return "", errors.New("error fetching siteUrl")
	}

	return *siteUrlRef, nil
}

func (p *Plugin) getOAuthConfig() (*oauth2.Config, error) {
	config := p.getConfiguration()

	clientID := config.OAuth2ClientID
	clientSecret := config.OAuth2ClientSecret
	clientAuthority := config.OAuth2Authority

	siteUrl, err := p.getSiteUrl()
	if err != nil {
		return nil, err
	}

	redirectUrl := fmt.Sprintf("%s/plugins/"+manifest.Id+"/oauth2/complete", siteUrl)

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Scopes: []string{
			"offline_access",
			"OnlineMeetings.ReadWrite",
		},
		Endpoint: microsoft.AzureADEndpoint(clientAuthority),
	}, nil
}

type UserInfo struct {
	Email string

	// OAuth Token, ttl 15 years
	OAuthToken *oauth2.Token

	// Mattermost userID
	UserID string

	// Remote userID
	RemoteID string
}

type AuthError struct {
	Message string `json:"message"`
	Err     error  `json:"err"`
}

func (ae *AuthError) Error() string {
	errorString, _ := json.Marshal(ae)
	return string(errorString)
}

func (p *Plugin) storeUserInfo(info *UserInfo) error {
	jsonInfo, err := json.Marshal(info)
	if err != nil {
		return err
	}

	if err := p.API.KVSet(tokenKey+info.UserID, jsonInfo); err != nil {
		return err
	}

	if err := p.API.KVSet(tokenKeyByRemoteID+info.RemoteID, jsonInfo); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) getUserInfo(userID string) (*UserInfo, error) {
	var userInfo UserInfo

	infoBytes, appErr := p.API.KVGet(tokenKey + userID)
	if appErr != nil || infoBytes == nil {
		return nil, errors.New("Must connect user account to Microsoft first.")
	}

	err := json.Unmarshal(infoBytes, &userInfo)
	if err != nil {
		return nil, errors.New("Unable to parse token.")
	}

	return &userInfo, nil
}

func (p *Plugin) authenticateAndFetchUser(userID, userEmail, channelID string) (*msgraph.User, *AuthError) {
	var user *msgraph.User
	var err error

	userInfo, apiErr := p.getUserInfo(userID)
	oauthMsg := fmt.Sprintf(
		oAuthMessage,
		*p.API.GetConfig().ServiceSettings.SiteURL, channelID)

	if apiErr != nil || userInfo == nil {
		return nil, &AuthError{Message: oauthMsg, Err: apiErr}
	}
	user, err = p.getUserWithToken(userInfo.OAuthToken)
	if err != nil || user == nil {
		return nil, &AuthError{Message: oauthMsg, Err: apiErr}
	}

	return user, nil
}

func (p *Plugin) disconnect(userID string) error {
	rawInfo, appErr := p.API.KVGet(tokenKey + userID)
	if appErr != nil {
		return appErr
	}

	var info UserInfo
	err := json.Unmarshal(rawInfo, &info)
	if err != nil {
		return err
	}

	errByMattermostID := p.API.KVDelete(tokenKey + userID)
	errByRemoteID := p.API.KVDelete(tokenKeyByRemoteID + info.RemoteID)
	if errByMattermostID != nil {
		return errByMattermostID
	}
	if errByRemoteID != nil {
		return errByRemoteID
	}
	return nil
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

func (p *Plugin) dm(userID string, message string) error {
	channel, err := p.API.GetDirectChannel(userID, p.botUserID)
	if err != nil {
		p.API.LogInfo("Couldn't get bot's DM channel", "user_id", userID)
		return err
	}

	post := &model.Post{
		Message:   message,
		ChannelId: channel.Id,
		UserId:    p.botUserID,
	}

	_, err = p.API.CreatePost(post)
	if err != nil {
		return err
	}
	return nil
}
