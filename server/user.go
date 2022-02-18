package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/pkg/errors"
	msgraph "github.com/yaegashi/msgraph.go/beta"
	"golang.org/x/oauth2"
)

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

func (c *Client) GetMe() (*msgraph.User, error) {
	ctx := context.Background()
	graphUser, err := c.builder.Me().Request().Get(ctx)
	if err != nil {
		c.api.LogError(errors.Wrap(err, "cannot get user").Error())
		return nil, err
	}

	if graphUser == nil {
		err = errors.New("empty user")
		c.api.LogError(errors.Wrap(err, "cannot get user").Error())
		return nil, err
	}

	return graphUser, nil
}

func (p *Plugin) StoreUserInfo(info *UserInfo) error {
	encryptionKey := p.getConfiguration().EncryptionKey
	if encryptionKey != "" {
		token := info.OAuthToken.AccessToken
		encryptedToken, err := encrypt([]byte(encryptionKey), token)
		if err != nil {
			return errors.Wrap(err, "error occurred while encrypting access token")
		}
		info.OAuthToken.AccessToken = encryptedToken
		p.API.LogDebug("encrypted user token", "encrypted", lastN(encryptedToken, 4), "token", lastN(token, 4))
	}

	jsonInfo, err := json.Marshal(info)
	if err != nil {
		return err
	}

	if err := p.API.KVSet(tokenKey+info.UserID, jsonInfo); err != nil {
		return err
	}

	if err := p.API.KVSet(tokenKeyByRemoteID+info.RemoteID, jsonInfo); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) GetUserInfo(userID string) (*UserInfo, error) {
	var userInfo UserInfo

	infoBytes, appErr := p.API.KVGet(tokenKey + userID)
	if appErr != nil || infoBytes == nil {
		return nil, errors.New("must connect user account to Microsoft first")
	}

	err := json.Unmarshal(infoBytes, &userInfo)
	if err != nil {
		return nil, errors.New("unable to parse token")
	}

	encryptionKey := p.getConfiguration().EncryptionKey
	if encryptionKey != "" {
		encryptedToken := userInfo.OAuthToken.AccessToken
		token, err := decrypt([]byte(encryptionKey), encryptedToken)
		if err != nil {
			p.API.LogWarn("Failed to decrypt access token", "error", err.Error())
			return nil, errors.Wrap(err, "failed to decrypt previously stored user access token")
		}

		userInfo.OAuthToken.AccessToken = token
		p.API.LogDebug("decrypted user token", "encrypted", lastN(encryptedToken, 4), "token", lastN(token, 4))
	}
	return &userInfo, nil
}

func (p *Plugin) RemoveUser(userID string) error {
	info, err := p.GetUserInfo(userID)
	if err != nil {
		return err
	}

	errByMattermostID := p.API.KVDelete(tokenKey + userID)
	errByRemoteID := p.API.KVDelete(tokenKeyByRemoteID + info.RemoteID)
	if errByMattermostID != nil {
		return errByMattermostID
	}
	if errByRemoteID != nil {
		return errByRemoteID
	}
	return nil
}

func encrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "could not create a cipher block, check key")
	}

	msg := pad([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", errors.Wrap(err, "readFull was unsuccessful, check buffer size")
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], msg)
	finalMsg := base64.URLEncoding.EncodeToString(ciphertext)
	return finalMsg, nil
}

func decrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "could not create a cipher block, check key")
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "", errors.Wrap(err, "could not decode the message")
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multiple of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := unpad(msg)
	if err != nil {
		return "", errors.Wrap(err, "unpad error, check key")
	}

	return string(unpadMsg), nil
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func lastN(s string, n int) string {
	out := []byte(s)
	if len(out) > n+3 {
		out = out[len(out)-n-3:]
	}
	for i := range out {
		if i < len(out)-n {
			out[i] = '*'
		}
	}
	return string(out)
}
