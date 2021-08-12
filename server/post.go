package main

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-msteams-meetings/server/remote"
	"github.com/mattermost/mattermost-plugin-msteams-meetings/server/store"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	msgraph "github.com/yaegashi/msgraph.go/beta"
)

func (p *Plugin) postWarning(creator *model.User, channelID string, userID string) (*model.Post, error) {
	channel, err := p.API.GetChannel(channelID)
	if err != nil {
		return nil, err
	}
	message := ""
	if channel.IsGroupOrDirect() {
		var members *model.ChannelMembers
		members, err = p.API.GetChannelMembers(channelID, 0, 100)
		if err != nil {
			return nil, err
		}

		if members != nil {
			p.API.LogDebug(fmt.Sprintf("%d members in channel %s", len(*members), channelID))
			membersCount := len(*members)
			message += "\n" + fmt.Sprintf("You are about a create a meeting in a channel with %d members", membersCount)
		}

	}

	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channelID,
		Message:   message,
		Type:      "custom_mstmeetings",
		Props: map[string]interface{}{
			"type":                     "custom_mstmeetings",
			"meeting_status":           postTypeDialogWarn,
			"meeting_personal":         true,
			"meeting_creator_username": creator.Username,
			"meeting_provider":         msteamsProviderName,
			"message":                  message,
		},
	}
	// postStr,  := json.Marshal(post)
	// p.API.LogDebug("post log", "post", string(postStr))
	return p.API.SendEphemeralPost(userID, post), nil
}

func (p *Plugin) postMeeting(creator *model.User, channelID string, topic string) (*model.Post, *msgraph.OnlineMeeting, error) {
	conf, err := p.getOAuthConfig()
	if err != nil {
		return nil, nil, err
	}
	userInfo, err := p.store.GetUserInfo(creator.Id)
	if err != nil {
		return nil, nil, err
	}

	attendees := []*store.UserInfo{}

	channel, appErr := p.API.GetChannel(channelID)
	if appErr != nil {
		return nil, nil, err
	}

	if channel.IsGroupOrDirect() {
		var members *model.ChannelMembers
		members, appErr = p.API.GetChannelMembers(channelID, 0, 100)
		if appErr != nil {
			return nil, nil, err
		}
		if members == nil {
			return nil, nil, errors.New("returned members is nil")
		}
		for _, member := range *members {
			var attendeeInfo *store.UserInfo
			attendeeInfo, err = p.store.GetUserInfo(member.UserId)
			if err != nil {
				continue
			}
			attendees = append(attendees, attendeeInfo)
		}
	}

	client := remote.NewClient(conf, userInfo.OAuthToken, p.API)

	meeting, err := client.CreateMeeting(userInfo, attendees)
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
			"meeting_provider":         msteamsProviderName,
		},
	}

	post, appErr = p.API.CreatePost(post)
	if appErr != nil {
		return nil, nil, appErr
	}

	return post, meeting, nil
}

func (p *Plugin) postConfirmCreateOrJoin(meetingURL string, channelID string, topic string, userID string, creatorName string, provider string) *model.Post {
	message := "There is another recent meeting created on this channel."
	if provider != msteamsProviderName {
		message = fmt.Sprintf("There is another recent meeting created on this channel with %s.", provider)
	}
	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channelID,
		Message:   message,
		Type:      "custom_mstmeetings",
		Props: map[string]interface{}{
			"type":                     "custom_mstmeetings",
			"meeting_link":             meetingURL,
			"meeting_status":           postTypeConfirm,
			"meeting_personal":         true,
			"meeting_topic":            topic,
			"meeting_creator_username": creatorName,
			"meeting_provider":         provider,
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
