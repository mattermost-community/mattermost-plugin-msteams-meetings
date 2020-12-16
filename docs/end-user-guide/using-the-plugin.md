## Connecting an MS Teams account to Mattermost

Use the `/mstmeetings connect` slash command to connect an MS Teams account to Mattermost.

## Starting a call

Start a call either by selecting the video icon in a Mattermost channel or by using the `/mstmeetings start` slash command. Every meeting you start creates a new meeting room in MS Teams. If you start two meetings less than 30 seconds apart you'll be prompted to confirm that you want to create the meeting.

## Disconnecting an account

You can disconnect your MS Teams account from Mattermost using the `/mstmeetings disconnect` slash command.

To connect another account to Mattermost, follow the configuration instructions again. If you are disconnecting in order to refresh the token, then use `/mstmeetings connect` to reconnect.
