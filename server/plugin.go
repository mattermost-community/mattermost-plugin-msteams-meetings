// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/mattermost/mattermost-plugin-msteams-meetings/server/store"

	"github.com/mattermost/mattermost-plugin-api/experimental/telemetry"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	botUserName    = "mstmeetings"
	botDisplayName = "MS Teams Meetings"
	botDescription = "Created by the MS Teams Meetings plugin."
)

// Plugin defines the plugin struct
type Plugin struct {
	plugin.MattermostPlugin

	// botUserID of the created bot account.
	botUserID string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	telemetryClient telemetry.Client
	tracker         telemetry.Tracker

	store store.Store
}

// OnActivate checks if the configurations is valid and ensures the bot account exists
func (p *Plugin) OnActivate() error {
	if !HasEnterpriseFeatures(p.API.GetConfig(), p.API.GetLicense()) {
		return errors.New("a valid Mattermost Enterprise license is required to use this plugin")
	}

	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		return err
	}

	if _, err := p.getSiteURL(); err != nil {
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

	p.store = store.Store{
		API: p.API,
	}

	p.telemetryClient, err = telemetry.NewRudderClient()
	if err != nil {
		p.API.LogWarn("telemetry client not started", "error", err.Error())
	}

	return nil
}

func (p *Plugin) OnDeactivate() error {
	if p.telemetryClient != nil {
		err := p.telemetryClient.Close()
		if err != nil {
			p.API.LogWarn("OnDeactivate: failed to close telemetryClient", "error", err.Error())
		}
	}

	return nil
}
