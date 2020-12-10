
# Mattermost MS Teams Meetings Plugin 

Start and join voice calls, video calls and use screen sharing with your team members via MS Teams Meetings.

Usage
-----

Once enabled, clicking the video icon in a Mattermost channel invites team members to join an MS Teams meeting, hosted using the credentials of the user who initiated the call.

![Screenshot](https://user-images.githubusercontent.com/177788/42196048-af54d2b8-7e30-11e8-80a0-5e160ae06f03.png)

## License

This repository is licensed under the [Mattermost Source Available License](LICENSE) and requires a valid Enterprise E20 license. See [Mattermost Source Available License](https://docs.mattermost.com/overview/faq.html#mattermost-source-available-license) to learn more.

## Development

This plugin contains both a server and web app portion. Read our documentation about the [Developer Workflow](https://developers.mattermost.com/extend/plugins/developer-workflow/) and [Developer Setup](https://developers.mattermost.com/extend/plugins/developer-setup/) for more information about developing and extending plugins.

### Server

Inside the `/server` directory, you will find the Go files that make up the server-side of the plugin. Within there, build the plugin like you would any other Go application.

### Web App

Inside the `/webapp` directory, you will find the JS and React files that make up the client-side of the plugin. Within there, modify files and components as necessary. Test your syntax by running `npm run build`.
