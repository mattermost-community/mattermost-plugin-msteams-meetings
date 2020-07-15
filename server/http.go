// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-mstelephony/server/remote"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	msgraph "github.com/yaegashi/msgraph.go/beta"
	"golang.org/x/oauth2"
)

const (
	postTypeStarted = "STARTED"
	postTypeEnded   = "ENDED"
	postTypeConfirm = "RECENTLY_CREATED"
)

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		http.Error(w, "This plugin is not configured.", http.StatusNotImplemented)
		return
	}

	switch path := r.URL.Path; path {
	case "/api/v1/meetings":
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
	}

	key := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)
	state := fmt.Sprintf("%v_%v", key, channelID)

	appErr := p.API.KVSet(key, []byte(state))
	if appErr != nil {
		http.Error(w, appErr.Error(), http.StatusInternalServerError)
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
	stateComponents := strings.Split(state, "_")

	if len(stateComponents) != stateLength {
		log.Printf("stateComponents: %v, state: %v", stateComponents, state)
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("%v_%v", stateComponents[0], stateComponents[1])

	var storedState []byte
	var appErr *model.AppError
	storedState, appErr = p.API.KVGet(key)
	if appErr != nil {
		http.Error(w, "missing stored state", http.StatusBadRequest)
		return
	}

	if string(storedState) != state {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	userID := stateComponents[1]
	channelID := stateComponents[2]

	p.API.KVDelete(state)

	if userID != authedUserID {
		http.Error(w, "Not authorized, incorrect user", http.StatusUnauthorized)
		return
	}

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		p.API.LogDebug(errors.Wrap(err, "error getting token").Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	remoteUser, err := p.getUserWithToken(tok)
	if err != nil {
		p.API.LogDebug(errors.Wrap(err, "error getting user").Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo := &UserInfo{
		UserID:     userID,
		OAuthToken: tok,
		Email:      *remoteUser.Mail,
		RemoteID:   *remoteUser.ID,
	}

	err = p.storeUserInfo(userInfo)
	if err != nil {
		p.API.LogDebug(errors.Wrap(err, "error storing the user info").Error())
		http.Error(w, "Unable to connect user to Microsoft", http.StatusInternalServerError)
		return
	}

	user, _ := p.API.GetUser(userID)

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
	w.Write([]byte(html))
}

type startMeetingRequest struct {
	ChannelID string `json:"channel_id"`
	Personal  bool   `json:"personal"`
	Topic     string `json:"topic"`
	MeetingID int    `json:"meeting_id"`
}

func (p *Plugin) postMeeting(creator *model.User, channelID string, topic string) (*model.Post, *msgraph.OnlineMeeting, error) {
	conf, err := p.getOAuthConfig()
	if err != nil {
		return nil, nil, err
	}
	userInfo, err := p.getUserInfo(creator.Id)
	if err != nil {
		return nil, nil, err
	}

	client := remote.NewClient(conf, userInfo.OAuthToken, p.API)

	graphUser, err := client.GetMe()
	if err != nil {
		return nil, nil, err
	}

	meeting, err := client.CreateMeeting(*graphUser.ID)
	if err != nil {
		return nil, nil, err
	}

	post := &model.Post{
		UserId:    creator.Id,
		ChannelId: channelID,
		Message:   fmt.Sprintf("Meeting started at [this link](%s).", *meeting.JoinURL),
		Type:      "custom_mstmeetings",
		Props: map[string]interface{}{
			"meeting_link":             *meeting.JoinURL,
			"meeting_status":           postTypeStarted,
			"meeting_personal":         true,
			"meeting_topic":            topic,
			"meeting_creator_username": creator.Username,
		},
	}

	post, appErr := p.API.CreatePost(post)
	if appErr != nil {
		return nil, nil, appErr
	}

	return post, meeting, nil
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
		recentMeeting, recentMeetingURL, creatorName, cpmErr := p.checkPreviousMessages(req.ChannelID)
		if cpmErr != nil {
			http.Error(w, cpmErr.Error(), cpmErr.StatusCode)
			return
		}

		if recentMeeting {
			_, err = w.Write([]byte(`{"meeting_url": ""}`))
			if err != nil {
				p.API.LogWarn("failed to write response", "error", err.Error())
			}
			p.postConfirm(recentMeetingURL, req.ChannelID, req.Topic, userID, creatorName)
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

	_, err = w.Write([]byte(fmt.Sprintf(`{"meeting_url": "%s"}`, *meeting.JoinURL)))
	if err != nil {
		p.API.LogWarn("failed to write response", "error", err.Error())
	}
}

func (p *Plugin) postConfirm(meetingURL string, channelID string, topic string, userID string, creatorName string) *model.Post {
	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channelID,
		Message:   "There is another recent meeting created on this channel.",
		Type:      "custom_mstmeetings",
		Props: map[string]interface{}{
			"type":                     "custom_mstmeetings",
			"meeting_link":             meetingURL,
			"meeting_status":           postTypeConfirm,
			"meeting_personal":         true,
			"meeting_topic":            topic,
			"meeting_creator_username": creatorName,
		},
	}

	return p.API.SendEphemeralPost(userID, post)
}

func (p *Plugin) postConnect(channelID string, userID string) *model.Post {
	oauthMsg := fmt.Sprintf(
		oAuthMessage,
		*p.API.GetConfig().ServiceSettings.SiteURL, channelID)

	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channelID,
		Message:   oauthMsg,
	}

	return p.API.SendEphemeralPost(userID, post)
}

func (p *Plugin) checkPreviousMessages(channelID string) (recentMeeting bool, meetingLink string, creatorName string, err *model.AppError) {
	var meetingTimeWindow int64 = 30 // 30 seconds

	postList, appErr := p.API.GetPostsSince(channelID, (time.Now().Unix()-meetingTimeWindow)*1000)
	if appErr != nil {
		return false, "", "", appErr
	}

	for _, post := range postList.ToSlice() {
		if meetingLink, ok := post.Props["meeting_link"]; ok {
			return true, meetingLink.(string), post.Props["meeting_creator_username"].(string), nil
		}
	}

	return false, "", "", nil
}
