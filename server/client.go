// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"context"

	"github.com/mattermost/mattermost-server/v6/plugin"
	msgraph "github.com/yaegashi/msgraph.go/beta"
	"golang.org/x/oauth2"
)

// Client represents a MSGraph API client
type Client struct {
	builder *msgraph.GraphServiceRequestBuilder
	api     plugin.API
}

// NewClient returns a new MSGraph API client.
func (p *Plugin) NewClient(conf *oauth2.Config, token *oauth2.Token) *Client {
	ctx := context.Background()
	httpClient := conf.Client(ctx, token)
	return &Client{
		builder: msgraph.NewClient(httpClient),
		api:     p.API,
	}
}
