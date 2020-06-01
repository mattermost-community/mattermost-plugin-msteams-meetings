// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	msgraph "github.com/yaegashi/msgraph.go/beta"
)

func (c *Client) CreateMeeting(userID string) (*msgraph.OnlineMeeting, error) {
	ctx := context.Background()
	start := time.Now()
	end := start.Add(1 * time.Hour)
	subject := "some subject"
	in := msgraph.OnlineMeeting{
		StartDateTime: &start,
		EndDateTime:   &end,
		Subject:       &subject,
	}
	out := msgraph.OnlineMeeting{}

	err := c.builder.Users().ID(userID).OnlineMeetings().Request().JSONRequest(ctx, http.MethodPost, "", &in, &out)
	if err != nil {
		return nil, errors.Wrap(err, "cannot creat meeting")
	}
	return &out, nil
}
