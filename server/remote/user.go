// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import (
	"context"

	"github.com/pkg/errors"
	msgraph "github.com/yaegashi/msgraph.go/beta"
)

func (c *Client) GetMe() (*msgraph.User, error) {
	ctx := context.Background()
	graphUser, err := c.builder.Me().Request().Get(ctx)
	if err != nil {
		c.api.LogError(errors.Wrap(err, "cannot get user").Error())
		return nil, err
	}
	return graphUser, nil
}
