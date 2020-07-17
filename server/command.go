package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	commandHelp = `* |/mstmeetings start| - Start an MS Teams meeting.
	* |/mstmeetings disconnect| - Disconnect from Mattermost`
)

func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "mstmeetings",
		DisplayName:      "MS Teams Meetings",
		Description:      "Integration with MS Teams Meetings.",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: start, disconnect",
		AutoCompleteHint: "[command]",
	}
}

func (p *Plugin) postCommandResponse(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	_ = p.API.SendEphemeralPost(args.UserId, post)
}

func (p *Plugin) executeCommand(c *plugin.Context, args *model.CommandArgs) (string, error) {
	split := strings.Fields(args.Command)
	command := split[0]
	action := ""

	if command != "/mstmeetings" {
		return fmt.Sprintf("Command '%s' is not /mstmeetings. Please try again.", command), nil
	}

	if len(split) > 1 {
		action = split[1]
	} else {
		return "Please specify an action for /mstmeetings command.", nil
	}

	userID := args.UserId
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return fmt.Sprintf("We could not retrieve user (userId: %v)", args.UserId), nil
	}

	if action == "start" {
		if _, appErr = p.API.GetChannelMember(args.ChannelId, userID); appErr != nil {
			return fmt.Sprintf("We could not get channel members (channelId: %v)", args.ChannelId), nil
		}

		recentMeeting, recentMeetingURL, creatorName, appErr := p.checkPreviousMessages(args.ChannelId)
		if appErr != nil {
			return fmt.Sprintf("Error checking previous messages"), nil
		}

		if recentMeeting {
			p.postConfirm(recentMeetingURL, args.ChannelId, "", userID, creatorName)
			return "", nil
		}

		_, authErr := p.authenticateAndFetchUser(userID, user.Email, args.ChannelId)
		if authErr != nil {
			return authErr.Message, authErr.Err
		}

		_, _, err := p.postMeeting(user, args.ChannelId, "")
		if err != nil {
			return "Failed to post message. Please try again.", nil
		}
		return "", nil
	}

	if action == "disconnect" {
		err := p.disconnect(userID)
		if err != nil {
			return "Failed to disconnect the user, err=" + err.Error(), nil
		}
		return "User disconnected from MS Teams Meetings.", nil
	}

	return fmt.Sprintf("Unknown action %v", action), nil
}

// ExecuteCommand is called when any registered by this plugin command is executed
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	msg, err := p.executeCommand(c, args)
	if err != nil {
		p.API.LogWarn("failed to execute command", "error", err.Error())
	}
	if msg != "" {
		p.postCommandResponse(args, msg)
	}
	return &model.CommandResponse{}, nil
}
