package store

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	stateLength = 3
)

func (s *Store) StoreState(userID string, extra string) (string, error) {
	key := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)
	state := fmt.Sprintf("%v_%v", key, extra)

	appErr := s.API.KVSet(key, []byte(state))
	if appErr != nil {
		return "", appErr
	}

	return state, nil
}

func (s *Store) GetState(key string) (string, error) {
	storedState, appErr := s.API.KVGet(key)
	if appErr != nil {
		return "", appErr
	}
	return string(storedState), nil
}

func (s *Store) DeleteState(key string) error {
	appErr := s.API.KVDelete(key)
	if appErr != nil {
		return appErr
	}
	return nil
}

func (s *Store) ParseState(state string) (key, userID, extra string, err error) {
	stateComponents := strings.Split(state, "_")

	if len(stateComponents) != stateLength {
		s.API.LogDebug("complete oauth, state mismatch", "stateComponents", fmt.Sprintf("%v", stateComponents), "state", state)
		return "", "", "", errors.New("status mismatch")
	}

	key = fmt.Sprintf("%v_%v", stateComponents[0], stateComponents[1])
	userID = stateComponents[1]
	extra = stateComponents[2]

	return key, userID, extra, nil
}
