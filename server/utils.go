package main

import (
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

func (p *Plugin) getSiteURL() (string, error) {
	siteURLRef := p.API.GetConfig().ServiceSettings.SiteURL
	if siteURLRef == nil || *siteURLRef == "" {
		return "", errors.New("error fetching siteUrl")
	}

	return *siteURLRef, nil
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

func (p *Plugin) dm(userID string, message string) error {
	channel, err := p.API.GetDirectChannel(userID, p.botUserID)
	if err != nil {
		p.API.LogInfo("couldn't get bot's DM channel", "user_id", userID, "bot_id", p.botUserID, "error", err.Error())
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
