// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-msteams-meetings/server/store"

	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	postTypeStarted    = "STARTED"
	postTypeDialogWarn = "WARN_DIALOG"
	postTypeConfirm    = "RECENTLY_CREATED"

	msteamsProviderName = "Microsoft Teams Meetings"
)

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		http.Error(w, "This plugin is not configured.", http.StatusNotImplemented)
		return
	}

	switch path := r.URL.Path; path {
	case "/api/v1/meetings/confirm":
		p.handleWarnAndConfirm(w, r)
	case "/api/v1/meetings/start":
		p.handleStartMeeting(w, r)
	case "/oauth2/connect":
		p.connectUser(w, r)
	case "/oauth2/complete":
		p.completeUserOAuth(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) connectUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	channelID := r.URL.Query().Get("channelID")
	if channelID == "" {
		http.Error(w, "channelID missing", http.StatusBadRequest)
		return
	}

	conf, err := p.getOAuthConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state, err := p.store.StoreState(userID, channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func (p *Plugin) completeUserOAuth(w http.ResponseWriter, r *http.Request) {
	authedUserID := r.Header.Get("Mattermost-User-ID")
	if authedUserID == "" {
		http.Error(w, "Not authorized, missing Mattermost user id", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	conf, err := p.getOAuthConfig()
	if err != nil {
		http.Error(w, "error in oauth config", http.StatusInternalServerError)
	}

	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")

	key, userID, channelID, err := p.store.ParseState(state)
	if err != nil {
		p.API.LogDebug("complete oauth, cannot parse state", "error", err.Error())
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	storedState, err := p.store.GetState(key)
	if err != nil {
		http.Error(w, "missing stored state", http.StatusBadRequest)
		return
	}

	if storedState != state {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	_ = p.store.DeleteState(key)

	if userID != authedUserID {
		http.Error(w, "Not authorized, incorrect user", http.StatusUnauthorized)
		return
	}

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		p.API.LogDebug("complete oauth, error getting token", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	remoteUser, err := p.getUserWithToken(tok)
	if err != nil {
		p.API.LogDebug("complete oauth, error getting user", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if remoteUser.Mail == nil {
		p.API.LogDebug("user has no mail")
		http.Error(w, "User has no mail. Please check the user is properly configured in Microsoft", http.StatusInternalServerError)
		return
	}

	if remoteUser.ID == nil {
		p.API.LogDebug("user has no ID")
		http.Error(w, "User has no ID. Please check the user is properly configured in Microsoft", http.StatusInternalServerError)
		return
	}

	if remoteUser.UserPrincipalName == nil {
		p.API.LogDebug("user has no UPN")
		http.Error(w, "User has no user principal name. Please check the user is properly configured in Microsoft", http.StatusInternalServerError)
		return
	}

	userInfo := &store.UserInfo{
		UserID:     userID,
		OAuthToken: tok,
		Email:      *remoteUser.Mail,
		RemoteID:   *remoteUser.ID,
		UPN:        *remoteUser.UserPrincipalName,
	}

	err = p.store.StoreUserInfo(userInfo)
	if err != nil {
		p.API.LogDebug("complete oauth, error storing the user info", "error", err.Error())
		http.Error(w, "Unable to connect user to Microsoft", http.StatusInternalServerError)
		return
	}

	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		p.API.LogError("complete oauth, error getting MM user", "error", appErr.Error())
		http.Error(w, appErr.Error(), http.StatusInternalServerError)
		return
	}

	p.trackConnect(userID)

	_, _, err = p.postMeeting(user, channelID, "")
	if err != nil {
		p.API.LogDebug(errors.Wrap(err, "error posting meeting").Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := `
<!DOCTYPE html>
<html>
	<head>
		<script>
			window.close();
		</script>
	</head>
	<body>
		<p>Completed connecting to Microsoft. Please close this window.</p>
	</body>
</html>
`

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(html))
}

type startMeetingRequest struct {
	ChannelID string `json:"channel_id"`
	Personal  bool   `json:"personal"`
	Topic     string `json:"topic"`
	MeetingID int    `json:"meeting_id"`
}

type confirmMeeting struct {
	ChannelID string `json:"channel_id"`
	Personal  bool   `json:"personal"`
}

func (p *Plugin) handleWarnAndConfirm(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")

	if userID == "" {
		p.API.LogWarn("handleWarnAndConfirm 1")
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	var req confirmMeeting
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		p.API.LogWarn("handleWarnAndConfirm 2", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		p.API.LogWarn("handleWarnAndConfirm 3")
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}
	p.postWarning(user, req.ChannelID, userID)
}

func (p *Plugin) handleStartMeeting(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	var req startMeetingRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	_, appErr = p.API.GetChannelMember(req.ChannelID, userID)
	if appErr != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.URL.Query().Get("force") == "" {
		recentMeeting, recentMeetingURL, creatorName, provider, cpmErr := p.checkPreviousMessages(req.ChannelID)
		if cpmErr != nil {
			http.Error(w, cpmErr.Error(), cpmErr.StatusCode)
			return
		}

		if recentMeeting {
			_, err = w.Write([]byte(`{"meeting_url": ""}`))
			if err != nil {
				p.API.LogWarn("failed to write response", "error", err.Error())
			}
			p.postConfirmCreateOrJoin(recentMeetingURL, req.ChannelID, req.Topic, userID, creatorName, provider)
			p.trackMeetingDuplication(userID)
			return
		}
	}

	_, authErr := p.authenticateAndFetchUser(userID, user.Email, req.ChannelID)
	if authErr != nil {
		if _, err = w.Write([]byte(`{"meeting_url": ""}`)); err != nil {
			p.API.LogWarn("failed to write response", "error", err.Error())
		}
		p.postConnect(req.ChannelID, userID)
		return
	}

	_, meeting, err := p.postMeeting(user, req.ChannelID, req.Topic)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p.trackMeetingStart(userID, telemetryStartSourceWebapp)
	if r.URL.Query().Get("force") != "" {
		p.trackMeetingForced(userID)
	}

	_, err = w.Write([]byte(fmt.Sprintf(`{"meeting_url": "%s"}`, *meeting.JoinURL)))
	if err != nil {
		p.API.LogWarn("failed to write response", "error", err.Error())
	}
}
