package main

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestEncryptUserData(t *testing.T) {
	exp, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z07:00")
	ui := UserInfo{
		Email: "test@test",
		OAuthToken: &oauth2.Token{
			AccessToken:  "access_t",
			TokenType:    "t_type",
			RefreshToken: "refresh_t",
			Expiry:       exp,
		},
		EncryptedOAuthToken: "to be wiped out",
		UserID:              "test",
		RemoteID:            "test-remote",
		UPN:                 "test-upn",
	}

	expected := ui
	expected.EncryptedOAuthToken = ""

	key := []byte("0123456789012345")
	data, err := ui.EncryptedJSON(key)
	require.NoError(t, err)
	require.Regexp(t,
		`\{"Email":"test@test","EncryptedOAuthToken":"[^"]+","UserID":"test","RemoteID":"test-remote","UPN":"test-upn"\}`,
		string(data))

	decrypted, err := DecryptUserInfo(data, key)
	require.NoError(t, err)
	require.EqualValues(t, &expected, decrypted)
}

func TestStoreUserInfo(t *testing.T) {
	tests := []struct {
		name           string
		kvSetUserErr   error
		kvSetRemoteErr error
		expectedErr    string
	}{
		{
			name:           "Error Saving UserID",
			kvSetUserErr:   &model.AppError{Message: "some error occurred while saving the user id"},
			kvSetRemoteErr: nil,
			expectedErr:    "some error occurred while saving the user id",
		},
		{
			name:           "Error Saving RemoteID",
			kvSetUserErr:   nil,
			kvSetRemoteErr: &model.AppError{Message: "some error occurred while saving the remote id"},
			expectedErr:    "some error occurred while saving the remote id",
		},
		{
			name:           "User Info stored successfully",
			kvSetUserErr:   nil,
			kvSetRemoteErr: nil,
			expectedErr:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &plugintest.API{}
			p := &Plugin{
				MattermostPlugin: plugin.MattermostPlugin{
					API: api,
				},
			}
			p.setConfiguration(&configuration{
				EncryptionKey: "demo_encrypt_key",
			})

			dummyInfo := &UserInfo{
				UserID:   "dummyUserID",
				RemoteID: "dummyRemoteID",
			}

			api.On("KVSet", "token_"+dummyInfo.UserID, mock.Anything).Return(tt.kvSetUserErr)
			if tt.kvSetUserErr == nil {
				api.On("KVSet", "tbyrid_"+dummyInfo.RemoteID, mock.Anything).Return(tt.kvSetRemoteErr)
			}

			responseErr := p.StoreUserInfo(dummyInfo)
			if tt.expectedErr == "" {
				require.NoError(t, responseErr)
			} else {
				require.Error(t, responseErr)
				require.Equal(t, tt.expectedErr, responseErr.Error())
			}
			api.AssertExpectations(t)
		})
	}
}

func TestResetAllOAuthTokens(t *testing.T) {
	tests := []struct {
		name           string
		kvDeleteAllErr error
		expectLogError bool
	}{
		{
			name:           "No Error",
			kvDeleteAllErr: nil,
			expectLogError: false,
		},
		{
			name:           "Error Deleting Tokens",
			kvDeleteAllErr: &model.AppError{Message: "error in deleting all oauth token"},
			expectLogError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &plugintest.API{}
			p := &Plugin{
				MattermostPlugin: plugin.MattermostPlugin{
					API: api,
				},
			}

			api.On("LogInfo", "OAuth2 configuration changed. Resetting all users' tokens, everyone will need to reconnect to MS Teams").Return(nil)
			api.On("KVDeleteAll").Return(tt.kvDeleteAllErr)

			if tt.expectLogError {
				api.On("LogError", "failed to reset user's OAuth2 tokens", "error", tt.kvDeleteAllErr.Error()).Return(nil)
			}

			p.resetAllOAuthTokens()
			api.AssertExpectations(t)
		})
	}
}
