# Mattermost MS Teams Meetings Plugin

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-msteams-meetings/master)](https://circleci.com/gh/mattermost/mattermost-plugin-msteams-meetings)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-msteams-meetings/master)](https://codecov.io/gh/mattermost/mattermost-plugin-msteams-meetings)
[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-msteams-meetings)](https://github.com/mattermost/mattermost-plugin-msteams-meetings/releases/latest)
[![HW](https://img.shields.io/github/issues/mattermost/mattermost-plugin-msteams-meetings/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/mattermost/mattermost-plugin-msteams-meetings/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

**Maintainer:** [@larkox](https://github.com/larkox)

Start and join voice calls, video calls, and use screen sharing with your team members via MS Teams Meetings.

## Admin guide

### Installation

The Mattermost MS Teams Meetings plugin is provided in the Mattermost Plugin Marketplace. Once enabled, selecting the video icon in a Mattermost channel invites team members to join an MS Teams meeting, hosted using the credentials of the user who initiated the call.

### Requirements

Mattermost Server v5.26+ is required.

#### Marketplace installation

1. In Mattermost, go to **Main Menu > Plugin Marketplace**.
2. Search for "MS Teams" or manually find the plugin from the list and select **Install**.
3. After the plugin is downloaded and installed, select **Configure**.

#### Manual Installation

If your server doesn't have access to the internet, you can download the latest [plugin binary release](https://github.com/mattermost/mattermost-plugin-msteams-meetings/releases) and upload it to your server via **System Console > Plugin Management**. The releases on this page are the same versions available on the Plugin Marketplace.

### Configuration

#### Step 1: Create a Mattermost App in Azure

1. Sign in to [the Azure portal](https://portal.azure.com) using an admin Azure account.
2. Navigate to [App Registrations](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps).
3. Select **New registration** at the top of the page.

    <img width="300" src="https://user-images.githubusercontent.com/6913320/76347903-be67f580-62dd-11ea-829e-236dd45865a8.png"/>

4. Fill out the form with the following values:

    - Name: **Mattermost MS Teams Meetings Plugin**
    - Supported account types: **Default value (Single tenant)**
    - Redirect URI: **https://(MM_SITE_URL)/plugins/com.mattermost.msteamsmeetings/oauth2/complete**. Replace `(MM_SITE_URL)` with your Mattermost server's URL.

5. Select **Register** to submit the form.

    <img width="500" src="https://user-images.githubusercontent.com/6913320/76348298-55cd4880-62de-11ea-8e0e-4ace3a8f8fcb.png"/>

6. Navigate to **Certificates & secrets** in the left pane.

    <img width="300" src="https://user-images.githubusercontent.com/6913320/76348833-3d116280-62df-11ea-8b13-d39a0a2f2024.png"/>

7. Select **New client secret > Add**, then copy the new secret in the bottom right corner of the screen. We'll use this value later in the Mattermost System Console.

    <img width="300" src="https://user-images.githubusercontent.com/6913320/76349025-9da09f80-62df-11ea-8c8f-0b39cad4597e.png"/>

8. Navigate to **API permissions** in the left pane.

    <img width="300" src="https://user-images.githubusercontent.com/6913320/76349582-a9d92c80-62e0-11ea-9414-5efd12c09b3f.png"/>

9. Select **Add a permission** and choose **Microsoft Graph** in the right pane.

    <img width="500" src="https://user-images.githubusercontent.com/6913320/76350226-c2961200-62e1-11ea-9080-19a9b75c2aee.png"/>

10. Select **Delegated permissions**, and scroll down to select the `OnlineMeetings.ReadWrite` permissions.

    <img width="300" src="https://user-images.githubusercontent.com/6913320/76350551-5a93fb80-62e2-11ea-8eb3-812735691af9.png"/>

11. Select **Add permissions** to submit the form.

    <img width="300" src="https://user-images.githubusercontent.com/6913320/80412303-abb07c80-889b-11ea-9640-7c2f264c790f.png"/>

12. Select **Grant admin consent for...** to grant the permissions for the application.

You're all set for configuration inside of the Azure portal.

#### Step 2: Configure plugin settings

1. Copy the **Client ID** and **Tenant ID** from the Azure portal.

    <img width="300" src="https://user-images.githubusercontent.com/6913320/76779336-9109c480-6781-11ea-8cde-4b79e5b2f3cd.png"/>

2. Go to **System Console > Plugins > MS Teams Meetings**.
3. Enter the following values in the fields provided:

    - `tenantID` - Copy from the Azure portal
    - `clientID` - Copy from the Azure portal
    - `Client Secret` - Copy from the Azure portal (generated in **Certificates & secrets** earlier in these instructions)

4. Choose **Save** to apply the configuration.

### Onboard users

When you’ve tested the plugin and confirmed it’s working, notify your team so they can get started. Copy and paste the text below, edit it to suit your requirements, and send it out.

> Hi team,

> The MS Teams Meetings plugin has been configured so you can use it for calls from within Mattermost. To get started, run the `/mstmeetings connect` slash command from any channel within Mattermost. Visit the documentation for more information.

## User guide

### Connect an MS Teams Account to Mattermost

Use the `/mstmeetings connect` slash command to connect an MS Teams account to Mattermost.

## Start a call

Start a call either by selecting the video icon in a Mattermost channel or by using the `/mstmeetings start` slash command. Every meeting you start creates a new meeting room in MS Teams. If you start two meetings less than 30 seconds apart you'll be prompted to confirm that you want to create the meeting.

## Disconnect an MS Teams account from Mattermost

Use the `/mstmeetings disconnect` slash command to disconnect an MS Teams account from Mattermost.

## Development

### Environment

This plugin contains both a server and web app portion. Read our documentation about the [Developer Workflow](https://developers.mattermost.com/extend/plugins/developer-workflow/) and [Developer Setup](https://developers.mattermost.com/extend/plugins/developer-setup/) for more information about developing and extending plugins.

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
