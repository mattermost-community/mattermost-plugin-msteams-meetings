package main

import (
	"errors"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/pluginapi/experimental/telemetry"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	msgraph "github.com/yaegashi/msgraph.go/beta"
)

type MockTracker struct {
	mock.Mock
}

func (m *MockTracker) TrackEvent(event string, properties map[string]interface{}) error {
	args := m.Called(event, properties)
	return args.Error(0)
}

func (m *MockTracker) TrackError(err error, properties map[string]interface{}) {
	m.Called(err, properties)
}

func (m *MockTracker) ReloadConfig(telemetry.TrackerConfig) {
	m.Called()
}

func (m *MockTracker) TrackUserEvent(userID string, event string, properties map[string]interface{}) error {
	args := m.Called(userID, event, properties)
	return args.Error(0)
}

type MockClient struct {
	mock.Mock
}

func (m *MockClient) GetMe() (*msgraph.User, error) {
	args := m.Called()
	return args.Get(0).(*msgraph.User), args.Error(1)
}

func (m *MockClient) CreateMeeting(creator *UserInfo, attendeesIDs []*UserInfo, subject string) (*msgraph.OnlineMeeting, error) {
	args := m.Called()
	return args.Get(0).(*msgraph.OnlineMeeting), args.Error(1)
}

func TestHandleConnect(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		commandArgs    *model.CommandArgs
		mockSetup      func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient)
		expectedOutput string
		expectError    bool
		expectedError  string
	}{
		{
			name:        "Too many parameters",
			args:        []string{"param1", "param2"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient) {
			},
			expectedOutput: tooManyParametersText,
			expectError:    false,
		},
		{
			name:        "Connect error",
			args:        []string{"param"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId", ChannelId: "demoChannelId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient) {
				api.On("KVGet", "token_demoUserId").Return(encryptedUserInfo, nil)
				api.On("KVSet", "msteamsmeetinguserstate_demoUserId", []byte("msteamsmeetinguserstate_demoUserId_demoChannelId_true")).Return(nil)
				api.On("GetConfig").Return(&model.Config{ServiceSettings: model.ServiceSettings{SiteURL: model.NewString("https://example.com")}})
				mockClient.On("GetMe").Return(&msgraph.User{}, errors.New("error getting user details"))
			},
			expectedOutput: "",
			expectError:    true,
			expectedError:  "error getting user details",
		},
		{
			name:        "Successful connection",
			args:        []string{"param"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId", ChannelId: "demoChannelId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient) {
				api.On("KVGet", "token_demoUserId").Return(encryptedUserInfo, nil)
				api.On("GetConfig").Return(&model.Config{ServiceSettings: model.ServiceSettings{SiteURL: model.NewString("https://example.com")}})
				mockClient.On("GetMe").Return(&msgraph.User{}, nil)
			},
			expectedOutput: "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &plugintest.API{}
			mockTracker := &MockTracker{}
			mockClient := &MockClient{}

			p := &Plugin{
				MattermostPlugin: plugin.MattermostPlugin{
					API:    api,
					Driver: &plugintest.Driver{},
				},
				tracker: mockTracker,
				client:  mockClient,
			}

			p.setConfiguration(&configuration{
				EncryptionKey:      "demo_encrypt_key",
				OAuth2ClientID:     "demo_oauth2_client_id",
				OAuth2ClientSecret: "demo_oauth2_client_secret",
				OAuth2Authority:    "demo_oauth2_authority",
			})

			userInfo := &UserInfo{
				Email:    "dummy@email.com",
				RemoteID: "demo_remote_id",
				UserID:   "dummy_user_id",
				UPN:      "dummy_upn",
			}

			encryptedUserInfo, err := userInfo.EncryptedJSON([]byte("demo_encrypt_key"))
			require.NoError(t, err)

			tt.mockSetup(api, encryptedUserInfo, mockTracker, mockClient)

			resp, err := p.handleConnect(tt.args, tt.commandArgs)
			if tt.expectError {
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Contains(t, resp, tt.expectedOutput)
			}

			api.AssertExpectations(t)
			mockTracker.AssertExpectations(t)
		})
	}
}

func TestHandleDisconnect(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		commandArgs    *model.CommandArgs
		mockSetup      func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker)
		expectedOutput string
	}{
		{
			name:           "Too many parameters",
			args:           []string{"param1", "param2"},
			commandArgs:    &model.CommandArgs{UserId: "demoUserId"},
			mockSetup:      func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker) {},
			expectedOutput: tooManyParametersText,
		},
		{
			name:        "Disconnect error",
			args:        []string{"param"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker) {
				api.On("KVGet", "token_demoUserId").Return(encryptedUserInfo, nil)
				api.On("KVDelete", "token_demoUserId").Return(&model.AppError{Message: "deletion error"})
				api.On("KVDelete", "tbyrid_demo_remote_id").Return(nil)
			},
			expectedOutput: "Failed to disconnect user, deletion error",
		},
		{
			name:        "Successful disconnection",
			args:        []string{"param"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker) {
				api.On("KVGet", "token_demoUserId").Return(encryptedUserInfo, nil)
				api.On("KVDelete", "token_demoUserId").Return(nil)
				api.On("KVDelete", "tbyrid_demo_remote_id").Return(nil)
				mockTracker.On("TrackUserEvent", "disconnect", "demoUserId", mock.Anything).Return(nil)
			},
			expectedOutput: "You have successfully disconnected from MS Teams Meetings.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &plugintest.API{}
			mockTracker := &MockTracker{}

			p := &Plugin{
				MattermostPlugin: plugin.MattermostPlugin{
					API:    api,
					Driver: &plugintest.Driver{},
				},
				tracker: mockTracker,
			}

			p.setConfiguration(&configuration{
				EncryptionKey: "demo_encrypt_key",
			})

			userInfo := &UserInfo{
				Email:    "dummy@email.com",
				RemoteID: "demo_remote_id",
				UserID:   "dummy_user_id",
				UPN:      "dummy_upn",
			}

			encryptedUserInfo, err := userInfo.EncryptedJSON([]byte("demo_encrypt_key"))
			require.NoError(t, err)

			tt.mockSetup(api, encryptedUserInfo, mockTracker)

			resp, err := p.handleDisconnect(tt.args, tt.commandArgs)
			require.NoError(t, err)
			require.Contains(t, resp, tt.expectedOutput)

			api.AssertExpectations(t)
			mockTracker.AssertExpectations(t)
		})
	}
}

func TestHandleStart(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		commandArgs    *model.CommandArgs
		mockSetup      func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient)
		expectedOutput string
		expectError    bool
		expectedError  string
	}{
		{
			name:        "Get User error",
			args:        []string{"param1", "param2"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient) {
				api.On("GetUser", "demoUserId").Return(nil, &model.AppError{Message: "error getting user for the userId"})
			},
			expectError:   true,
			expectedError: "error getting user for the userId",
		},
		{
			name:        "Get Channel error",
			args:        []string{"param1", "param2"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId", ChannelId: "demoChannelId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient) {
				api.On("GetUser", "demoUserId").Return(&model.User{Id: "demoUserId"}, nil)
				api.On("GetChannelMember", "demoChannelId", "demoUserId").Return(nil, &model.AppError{Message: "error getting channel for the channelId"})
			},
			expectError:   true,
			expectedError: "error getting channel for the channelId",
		},
		{
			name:        "Check Previous Message error",
			args:        []string{"param1", "param2"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId", ChannelId: "demoChannelId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient) {
				api.On("GetUser", "demoUserId").Return(&model.User{Id: "demoUserId"}, nil)
				api.On("GetChannelMember", "demoChannelId", "demoUserId").Return(&model.ChannelMember{ChannelId: "demoChannelId"}, nil)
				api.On("GetPostsSince", "demoChannelId", (time.Now().Unix()-30)*1000).Return(nil, &model.AppError{Message: "error getting previous post for channel"})
			},
			expectError:   true,
			expectedError: "error getting previous post for channel",
		},
		{
			name:        "Recent meeting exists",
			args:        []string{"param1", "param2"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId", ChannelId: "demoChannelId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient) {
				postList := &model.PostList{
					Order: []string{"post1"},
					Posts: map[string]*model.Post{
						"post1": {
							Id:        "post1",
							ChannelId: "demoChannelId",
							CreateAt:  time.Now().UnixNano() / int64(time.Millisecond),
							UpdateAt:  time.Now().UnixNano() / int64(time.Millisecond),
							Props: map[string]interface{}{
								"meeting_provider":         "meetingProvider",
								"meeting_link":             "meetingLink",
								"meeting_creator_username": "creatorName",
							},
						},
					},
					NextPostId:                "",
					PrevPostId:                "",
					HasNext:                   false,
					FirstInaccessiblePostTime: 0,
				}
				api.On("GetUser", "demoUserId").Return(&model.User{Id: "demoUserId"}, nil)
				api.On("GetChannelMember", "demoChannelId", "demoUserId").Return(&model.ChannelMember{ChannelId: "demoChannelId"}, nil)
				api.On("GetPostsSince", "demoChannelId", (time.Now().Unix()-30)*1000).Return(postList, nil)
				api.On("SendEphemeralPost", "demoUserId", mock.Anything).Return(&model.Post{})
				mockTracker.On("TrackUserEvent", mock.Anything, "demoUserId", mock.Anything).Return(nil)
			},
			expectError:    false,
			expectedOutput: "",
		},
		{
			name:        "Authentication error",
			args:        []string{"param1", "param2"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId", ChannelId: "demoChannelId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient) {
				api.On("GetUser", "demoUserId").Return(&model.User{Id: "demoUserId"}, nil)
				api.On("GetChannelMember", "demoChannelId", "demoUserId").Return(&model.ChannelMember{ChannelId: "demoChannelId"}, nil)
				postList := &model.PostList{
					Order: []string{"post1"},
					Posts: map[string]*model.Post{
						"post1": {
							Id:        "post1",
							ChannelId: "demoChannelId",
							CreateAt:  time.Now().UnixNano() / int64(time.Millisecond),
							UpdateAt:  time.Now().UnixNano() / int64(time.Millisecond),
						},
					},
					NextPostId:                "",
					PrevPostId:                "",
					HasNext:                   false,
					FirstInaccessiblePostTime: 0,
				}
				api.On("GetPostsSince", "demoChannelId", (time.Now().Unix()-30)*1000).Return(postList, nil)
				api.On("KVGet", "token_demoUserId").Return(nil, &model.AppError{Message: "deletion error"})
				api.On("GetConfig").Return(&model.Config{ServiceSettings: model.ServiceSettings{SiteURL: model.NewString("https://example.com")}})
				api.On("KVSet", "msteamsmeetinguserstate_demoUserId", []byte("msteamsmeetinguserstate_demoUserId_demoChannelId_false")).Return(nil)
			},
			expectError:   true,
			expectedError: "Your Mattermost account is not connected to any Microsoft Teams account",
		},
		{
			name:        "Meeting started successfully",
			args:        []string{"param1", "param2"},
			commandArgs: &model.CommandArgs{UserId: "demoUserId", ChannelId: "demoChannelId"},
			mockSetup: func(api *plugintest.API, encryptedUserInfo []byte, mockTracker *MockTracker, mockClient *MockClient) {
				api.On("GetUser", "demoUserId").Return(&model.User{Id: "demoUserId"}, nil)
				api.On("GetChannelMember", "demoChannelId", "demoUserId").Return(&model.ChannelMember{ChannelId: "demoChannelId"}, nil)

				postList := &model.PostList{
					Order: []string{"post1"},
					Posts: map[string]*model.Post{
						"post1": {
							Id:        "post1",
							ChannelId: "demoChannelId",
							CreateAt:  time.Now().UnixNano() / int64(time.Millisecond),
							UpdateAt:  time.Now().UnixNano() / int64(time.Millisecond),
						},
					},
					NextPostId:                "",
					PrevPostId:                "",
					HasNext:                   false,
					FirstInaccessiblePostTime: 0,
				}

				joinUrl := "demoJoinURL"

				api.On("GetPostsSince", "demoChannelId", (time.Now().Unix()-30)*1000).Return(postList, nil)
				api.On("KVGet", "token_demoUserId").Return(encryptedUserInfo, nil)
				api.On("GetConfig").Return(&model.Config{ServiceSettings: model.ServiceSettings{SiteURL: model.NewString("https://example.com")}})
				api.On("HasPermissionToChannel", "demoUserId", "demoChannelId", model.PermissionCreatePost).Return(true)
				api.On("GetChannel", "demoChannelId").Return(&model.Channel{Id: "demoChannelId", Type: model.ChannelTypeOpen}, nil)
				api.On("CreatePost", mock.Anything).Return(&model.Post{Id: "demoPostId"}, nil)
				mockClient.On("GetMe").Return(&msgraph.User{}, nil)
				mockClient.On("CreateMeeting", mock.Anything, mock.Anything, mock.Anything).Return(&msgraph.OnlineMeeting{JoinURL: &joinUrl}, nil)
				mockTracker.On("TrackUserEvent", "meeting_started", "demoUserId", mock.Anything).Return(nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &plugintest.API{}
			mockTracker := &MockTracker{}
			mockClient := &MockClient{}

			p := &Plugin{
				MattermostPlugin: plugin.MattermostPlugin{
					API: api,
				},
				tracker: mockTracker,
				client:  mockClient,
			}

			p.setConfiguration(&configuration{
				EncryptionKey: "demo_encrypt_key",
			})

			userInfo := &UserInfo{
				Email:    "dummy@email.com",
				RemoteID: "demo_remote_id",
				UserID:   "dummy_user_id",
				UPN:      "dummy_upn",
			}

			encryptedUserInfo, err := userInfo.EncryptedJSON([]byte("demo_encrypt_key"))
			require.NoError(t, err)

			tt.mockSetup(api, encryptedUserInfo, mockTracker, mockClient)

			resp, err := p.handleStart(tt.args, tt.commandArgs)
			if tt.expectError {
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Contains(t, resp, tt.expectedOutput)
			}

			api.AssertExpectations(t)
			mockTracker.AssertExpectations(t)
		})
	}
}

func TestGetHelpText(t *testing.T) {
	p := &Plugin{}
	expected := "###### Mattermost MS Teams Meetings Plugin - Slash Command Help\n" +
		"* `/mstmeetings start` - Start an MS Teams meeting. \n" +
		"* `/mstmeetings connect` - Connect to MS Teams meeting. \n" +
		"* `/mstmeetings disconnect` - Disconnect your Mattermost account from MS Teams. \n" +
		"* `/mstmeetings help` - Display this help text."

	actual := p.getHelpText()
	require.Equal(t, expected, actual)
}
