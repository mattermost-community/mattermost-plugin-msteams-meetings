# Mattermost MS Teams Meetings Plugin

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-msteams-meetings/master)](https://circleci.com/gh/mattermost/mattermost-plugin-msteams-meetings)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-msteams-meetings/master)](https://codecov.io/gh/mattermost/mattermost-plugin-msteams-meetings)
[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-msteams-meetings)](https://github.com/mattermost/mattermost-plugin-msteams-meetings/releases/latest)
[![HW](https://img.shields.io/github/issues/mattermost/mattermost-plugin-msteams-meetings/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/mattermost/mattermost-plugin-msteams-meetings/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

Start and join voice calls, video calls, and use screen sharing with your team members in Microsoft Teams Meetings.

## Admin guide

### Requirements

Mattermost Server v5.26+ is required.

### Installation

From Mattermost v10, this plugin is pre-packaged with the Mattermost Server.

If your Mattermost deployment is on a release prior to v10, download the latest [plugin binary release](https://github.com/mattermost/mattermost-plugin-msteams-meetings/releases), and upload it to your server via **System Console > Plugin Management**.

Once enabled, selecting the video icon in a Mattermost channel invites team members to join an MS Teams meeting, hosted using the credentials of the user who initiated the call.

### Configuration, Setup, and Usage

See the Mattermost Product Documentation for details on [setting up](https://docs.mattermost.com/integrate/microsoft-teams-meetings-interoperability.html#setup), [configuring](https://docs.mattermost.com/integrate/microsoft-teams-meetings-interoperability.html#enable-and-configure-the-microsoft-teams-meetings-integration-in-mattermost), and [using](https://docs.mattermost.com/integrate/microsoft-teams-meetings-interoperability.html#usage) the Mattermost for Microsoft Teams Meetings integration.

## Development

### Environment

This plugin contains both a server and web app portion. Read our documentation about the [Developer Workflow](https://developers.mattermost.com/integrate/plugins/developer-workflow/) and [Developer Setup](https://developers.mattermost.com/integrate/plugins/developer-setup/) for more information about developing and extending plugins.

#### Server

Inside the `/server` directory, you will find the Go files that make up the server-side of the plugin. Within that directory, build the plugin like you would any other Go application.

#### Web app

Inside the `/webapp` directory, you will find the JavaScript files that make up the client-side of the plugin. Within that directory, modify files and components as necessary. Test your syntax by running `npm run build`.

### Deploying to a Mattermost local server

It's on the [Developer setup](https://developers.mattermost.com/integrate/plugins/developer-setup/#deploy-with-local-mode), but keep in mind It's necessary to enable PluginUploads.

Then you need to set these envvars to be able to upload the plugin:

```bash
export MM_SERVICESETTINGS_SITEURL=https://localhost:8065/   # Or other if needed
export MM_ADMIN_USERNAME=<MYUSERNAME>
export MM_ADMIN_PASSWORD=<MYPASSWORD>
```

Then run `make deploy` or `MM_DEBUG=1 make deploy` in case you want debugging from the root of the `mattermost-plugin-msteams-meetings` project.

## Contact us

- For questions, suggestions, and help, visit the [Plugin: Microsoft Teams Meetings](https://community.mattermost.com/core/channels/plugin-microsoft-teams-meetings) on our Community server.
- To report a bug, please [open an issue](https://github.com/mattermost/mattermost-plugin-msteams-meetings/issues).
