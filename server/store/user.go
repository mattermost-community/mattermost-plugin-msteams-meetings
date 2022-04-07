package store

import (
	"encoding/json"
	"errors"

	"github.com/mattermost/mattermost-server/v5/plugin"
	"golang.org/x/oauth2"
)

type Store struct {
	API plugin.API
}

// UserInfo defines the information we store from each user
type UserInfo struct {
	Email      string
	OAuthToken *oauth2.Token
	// Mattermost userID
	UserID string
	// Remote userID
	RemoteID string
	// Remote UPN
	UPN string
}

const (
	tokenKey           = "token_"
	tokenKeyByRemoteID = "tbyrid_"
)

func (s *Store) StoreUserInfo(info *UserInfo) error {
	jsonInfo, err := json.Marshal(info)
	if err != nil {
		return err
	}

	if err := s.API.KVSet(tokenKey+info.UserID, jsonInfo); err != nil {
		return err
	}

	if err := s.API.KVSet(tokenKeyByRemoteID+info.RemoteID, jsonInfo); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetUserInfo(userID string) (*UserInfo, error) {
	var userInfo UserInfo

	infoBytes, appErr := s.API.KVGet(tokenKey + userID)
	if appErr != nil || infoBytes == nil {
		return nil, errors.New("Connect the user account to Microsoft Teams.")
	}

	err := json.Unmarshal(infoBytes, &userInfo)
	if err != nil {
		return nil, errors.New("unable to parse token")
	}

	return &userInfo, nil
}

func (s *Store) RemoveUser(userID string) error {
	info, err := s.GetUserInfo(userID)
	if err != nil {
		return err
	}

	errByMattermostID := s.API.KVDelete(tokenKey + userID)
	errByRemoteID := s.API.KVDelete(tokenKeyByRemoteID + info.RemoteID)
	if errByMattermostID != nil {
		return errByMattermostID
	}
	if errByRemoteID != nil {
		return errByRemoteID
	}
	return nil
}
