package main

import (
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
)

func (p *Plugin) getSiteURL() (string, error) {
	siteURLRef := p.API.GetConfig().ServiceSettings.SiteURL
	if siteURLRef == nil || *siteURLRef == "" {
		return "", errors.New("error fetching siteUrl")
	}

	return *siteURLRef, nil
}

func (p *Plugin) checkPreviousMessages(channelID string) (recentMeeting bool, meetingLink string, creatorName string, provider string, err *model.AppError) {
	var meetingTimeWindow int64 = 30 // 30 seconds

	postList, appErr := p.API.GetPostsSince(channelID, (time.Now().Unix()-meetingTimeWindow)*1000)
	if appErr != nil {
		return false, "", "", "", appErr
	}

	for _, post := range postList.ToSlice() {
		meetingProvider := getString("meeting_provider", post.Props)
		if meetingProvider == "" {
			continue
		}

		meetingLink := getString("meeting_link", post.Props)
		if meetingLink == "" {
			continue
		}

		creator := getString("meeting_creator_username", post.Props)

		return true, meetingLink, creator, meetingProvider, nil
	}

	return false, "", "", "", nil
}

func getString(key string, props model.StringInterface) string {
	value := ""
	if valueInterface, ok := props[key]; ok {
		if valueString, ok := valueInterface.(string); ok {
			value = valueString
		}
	}
	return value
}
