package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStoreState(t *testing.T) {
	testCases := []struct {
		name            string
		userID         string
		channelID      string
		justConnect     bool
		returnError     error
		expectError     bool
		expectedState   string
		expectedErrMsg  string
	}{
		{
			name:          "Store state successful",
			userID:       "userID1",
			channelID:    "channelID1",
			justConnect:   true,
			returnError:   nil,
			expectError:   false,
			expectedState: "msteamsmeetinguserstate_userID1_channelID1_true",
		},
		{
			name:          "Error occurred while storing state",
			userID:       "userID2",
			channelID:    "channelID2",
			justConnect:   false,
			returnError:   &model.AppError{Message: "error occurred while storing state"},
			expectError:   true,
			expectedErrMsg: "error occurred while storing state",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := &plugintest.API{}
			p := &Plugin{
				MattermostPlugin: plugin.MattermostPlugin{
					API: api,
				},
			}

			api.On("KVSet", mock.Anything, mock.Anything).Return(tc.returnError)

			state, err := p.StoreState(tc.userID, tc.channelID, tc.justConnect)

			if tc.expectError {
				require.Error(t, err)
				require.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedState, state)
			}
		})
	}
}

func TestGetState(t *testing.T) {
	testCases := []struct {
		name           string
		key            string
		returnValue    []byte
		returnError    error
		expectedState  string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:          "Valid state retrieved",
			key:           "dummyKey",
			returnValue:   []byte("dummyState"),
			returnError:   nil,
			expectedState: "dummyState",
			expectError:   false,
		},
		{
			name:           "Error occurred while getting stored state",
			key:            "dummyKey",
			returnValue:    []byte(""),
			returnError:    &model.AppError{Message: "error occurred while getting stored state"},
			expectedState:  "",
			expectError:    true,
			expectedErrMsg: "error occurred while getting stored state",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := &plugintest.API{}
			p := &Plugin{
				MattermostPlugin: plugin.MattermostPlugin{
					API: api,
				},
			}

			api.On("KVGet", tc.key).Return(tc.returnValue, tc.returnError)

			state, err := p.GetState(tc.key)
			if tc.expectError {
				require.Error(t, err)
				require.Equal(t, tc.expectedState, state)
				require.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedState, state)
			}
		})
	}
}

func TestDeleteState(t *testing.T) {
	testCases := []struct {
		name           string
		key            string
		returnError    error
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "Delete state successful",
			key:         "dummyKey",
			returnError: nil,
			expectError: false,
		},
		{
			name:           "Error occurred while deleting state",
			key:            "dummyKey",
			returnError:    &model.AppError{Message: "error occurred while deleting state"},
			expectError:    true,
			expectedErrMsg: "error occurred while deleting state",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := &plugintest.API{}
			p := &Plugin{
				MattermostPlugin: plugin.MattermostPlugin{
					API: api,
				},
			}

			api.On("KVDelete", tc.key).Return(tc.returnError)

			err := p.DeleteState(tc.key)
			if tc.expectError {
				require.Error(t, err)
				require.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseState(t *testing.T) {
	testCases := []struct {
		name                string
		state               string
		expectedKey         string
		expectedUserID      string
		expectedChannelID   string
		expectedJustConnect bool
		expectError         bool
		expectedErrMsg      string
	}{
		{
			name:                "Parse state successful",
			state:               "key1_userID1_channelID1_true",
			expectedKey:         "key1_userID1",
			expectedUserID:      "userID1",
			expectedChannelID:   "channelID1",
			expectedJustConnect: true,
			expectError:         false,
		},
		{
			name:                "State length mismatch",
			state:               "key1_userID1_channelID1",
			expectedKey:         "",
			expectedUserID:      "",
			expectedChannelID:   "",
			expectedJustConnect: false,
			expectError:         true,
			expectedErrMsg:      "status mismatch",
		},
		{
			name:                "Invalid state format",
			state:               "key1_userID1",
			expectedKey:         "",
			expectedUserID:      "",
			expectedChannelID:   "",
			expectedJustConnect: false,
			expectError:         true,
			expectedErrMsg:      "status mismatch",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := &plugintest.API{}
			p := &Plugin{
				MattermostPlugin: plugin.MattermostPlugin{
					API: api,
				},
			}

			// Set up the mock to expect LogDebug calls for error cases
			if tc.expectError {
				api.On("LogDebug", "complete oauth, state mismatch", "stateComponents", mock.Anything, "state", tc.state).Return()
			}

			key, userID, channelID, justConnect, err := p.ParseState(tc.state)

			if tc.expectError {
				require.Error(t, err)
				require.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedKey, key)
				require.Equal(t, tc.expectedUserID, userID)
				require.Equal(t, tc.expectedChannelID, channelID)
				require.Equal(t, tc.expectedJustConnect, justConnect)
			}
		})
	}
}
