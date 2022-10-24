// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	msgraph "github.com/yaegashi/msgraph.go/beta"
)

func (c *Client) CreateMeeting(creator *UserInfo, attendeesIDs []*UserInfo) (*msgraph.OnlineMeeting, error) {
	ctx := context.Background()
	start := time.Now()
	end := start.Add(1 * time.Hour)
	subject := "Mattermost Meeting"
	attendees := []msgraph.MeetingParticipantInfo{}
	for _, attendee := range attendeesIDs {
		attendees = append(attendees, msgraph.MeetingParticipantInfo{
			Identity: &msgraph.IdentitySet{
				User: &msgraph.Identity{
					ID: &attendee.RemoteID,
				},
			},
			Upn: &attendee.UPN,
		})
	}

	in := msgraph.OnlineMeeting{
		StartDateTime: &start,
		EndDateTime:   &end,
		Subject:       &subject,
		Participants: &msgraph.MeetingParticipants{
			Organizer: &msgraph.MeetingParticipantInfo{
				Identity: &msgraph.IdentitySet{
					User: &msgraph.Identity{
						ID: &creator.RemoteID,
					},
				},
				Upn: &creator.UPN,
			},
			Attendees: attendees,
		},
	}
	out := msgraph.OnlineMeeting{}

	err := c.builder.Users().ID(creator.RemoteID).OnlineMeetings().Request().JSONRequest(ctx, http.MethodPost, "", &in, &out)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create meeting")
	}
	return &out, nil
}
