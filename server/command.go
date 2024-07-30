package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/experimental/command"
	"github.com/pkg/errors"
)

const (
	commandHelp = "###### Mattermost MS Teams Meetings Plugin - Slash Command Help\n" +
		"* |/mstmeetings start| - Start an MS Teams meeting. \n" +
		"* |/mstmeetings connect| - Connect to MS Teams meeting. \n" +
		"* |/mstmeetings disconnect| - Disconnect your Mattermost account from MS Teams. \n" +
		"* |/mstmeetings help| - Display this help text."
	tooManyParametersText = "Too many parameters."
)

func getCommand(client *pluginapi.Client) *model.Command {
	iconData, err := command.GetIconData(&client.System, "assets/profile.svg")
	if err != nil {
		client.Log.Warn("Error getting icon data", "err", err.Error())
	}

	return &model.Command{
		Trigger:              "mstmeetings",
		DisplayName:          "MS Teams Meetings",
		Description:          "Integration with MS Teams Meetings.",
		AutoComplete:         true,
		AutoCompleteDesc:     "Available commands: start, connect, disconnect, help",
		AutoCompleteHint:     "[command]",
		AutocompleteIconData: iconData,
		AutocompleteData:     getAutocompleteData(),
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

func getAutocompleteData() *model.AutocompleteData {
	command := model.NewAutocompleteData("mstmeetings", "[command]",
		"Available commands: start, disconnect, help")

	start := model.NewAutocompleteData("start", "", "Start an MS Teams meeting")
	command.AddCommand(start)

	connect := model.NewAutocompleteData("connect", "",
		"Connect your Mattermost account to MS Teams")
	command.AddCommand(connect)

	disconnect := model.NewAutocompleteData("disconnect", "",
		"Disconnect your Mattermost account from MS Teams")
	command.AddCommand(disconnect)

	help := model.NewAutocompleteData("help", "", "Display usage information")
	command.AddCommand(help)

	return command
}

func (p *Plugin) executeCommand(_ *plugin.Context, args *model.CommandArgs) (string, error) {
	split := strings.Fields(args.Command)
	command := split[0]
	action := ""

	if command != "/mstmeetings" {
		return fmt.Sprintf("Command '%s' is not /mstmeetings. Please try again.", command), nil
	}

	if len(split) > 1 {
		action = split[1]
	} else {
		return p.handleHelp()
	}

	switch action {
	case "start":
		return p.handleStart(split[1:], args)
	case "connect":
		return p.handleConnect(split[1:], args)
	case "disconnect":
		return p.handleDisconnect(split[1:], args)
	case "help":
		return p.handleHelp()
	}

	return fmt.Sprintf("Unknown action `%v`.\n%s", action, p.getHelpText()), nil
}

func (p *Plugin) getHelpText() string {
	return strings.ReplaceAll(commandHelp, "|", "`")
}

func (p *Plugin) handleHelp() (string, error) {
	return p.getHelpText(), nil
}

func (p *Plugin) handleStart(args []string, extra *model.CommandArgs) (string, error) {
	topic := ""
	if len(args) > 1 {
		topic = strings.Join(args[1:], " ")
	}
	userID := extra.UserId
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return "Cannot get user.", errors.Wrap(appErr, "cannot get user")
	}

	if _, appErr = p.API.GetChannelMember(extra.ChannelId, userID); appErr != nil {
		return "We could not get channel members.", errors.Wrap(appErr, "cannot get channel member")
	}

	recentMeeting, recentMeetingURL, creatorName, provider, appErr := p.checkPreviousMessages(extra.ChannelId)
	if appErr != nil {
		return "Error checking previous messages.", errors.Wrap(appErr, "cannot check previous messages")
	}

	if recentMeeting {
		p.postConfirmCreateOrJoin(recentMeetingURL, extra.ChannelId, topic, userID, creatorName, provider)
		p.trackMeetingDuplication(extra.UserId)
		return "", nil
	}

	_, authErr := p.authenticateAndFetchUser(userID, extra.ChannelId)
	if authErr != nil {
		return authErr.Message, authErr.Err
	}

	_, _, err := p.postMeeting(user, extra.ChannelId, topic)
	if err != nil {
		return "Failed to post message. Please try again.", errors.Wrap(err, "cannot post message")
	}

	p.trackMeetingStart(extra.UserId, telemetryStartSourceCommand)
	return "", nil
}

func (p *Plugin) handleConnect(args []string, extra *model.CommandArgs) (string, error) {
	if len(args) > 1 {
		return tooManyParametersText, nil
	}

	msUser, authErr := p.authenticateAndFetchUser(extra.UserId, extra.ChannelId)
	if authErr != nil {
		return authErr.Message, authErr.Err
	}

	if msUser != nil {
		return "User already connected to MS Teams Meetings", nil
	}

	return "", nil
}

func (p *Plugin) handleDisconnect(args []string, extra *model.CommandArgs) (string, error) {
	if len(args) > 1 {
		return tooManyParametersText, nil
	}
	err := p.disconnect(extra.UserId)
	if err != nil {
		return fmt.Sprintf("Failed to disconnect user, %s", err.Error()), nil
	}

	p.trackDisconnect(extra.UserId)
	return "User disconnected from MS Teams Meetings.", nil
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
