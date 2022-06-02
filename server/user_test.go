package main

import (
	"testing"
	"time"

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
