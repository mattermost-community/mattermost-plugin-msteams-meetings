package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

const (
	stateLength = 3
)

func (p *Plugin) StoreState(userID string, extra string) (string, error) {
	key := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)
	state := fmt.Sprintf("%v_%v", key, extra)

	appErr := p.API.KVSet(key, []byte(state))
	if appErr != nil {
		return "", appErr
	}

	return state, nil
}

func (p *Plugin) GetState(key string) (string, error) {
	storedState, appErr := p.API.KVGet(key)
	if appErr != nil {
		return "", appErr
	}
	return string(storedState), nil
}

func (p *Plugin) DeleteState(key string) error {
	appErr := p.API.KVDelete(key)
	if appErr != nil {
		return appErr
	}
	return nil
}

func (p *Plugin) ParseState(state string) (key, userID, extra string, err error) {
	stateComponents := strings.Split(state, "_")

	if len(stateComponents) != stateLength {
		p.API.LogDebug("complete oauth, state mismatch", "stateComponents", fmt.Sprintf("%v", stateComponents), "state", state)
		return "", "", "", errors.New("status mismatch")
	}

	key = fmt.Sprintf("%v_%v", stateComponents[0], stateComponents[1])
	userID = stateComponents[1]
	extra = stateComponents[2]

	return key, userID, extra, nil
}
