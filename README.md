# Mattermost MS Teams Meetings Plugin 

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-msteams-meetings/master)](https://circleci.com/gh/mattermost/mattermost-plugin-msteams-meetings)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-msteams-meetings/master)](https://codecov.io/gh/mattermost/mattermost-plugin-msteams-meetings)
[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-msteams-meetings)](https://github.com/mattermost/mattermost-plugin-msteams-meetings/releases/latest)
[![HW](https://img.shields.io/github/issues/mattermost/mattermost-plugin-msteams-meetings/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/mattermost/mattermost-plugin-msteams-meetings/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)


**Maintainer:** [@larkox](https://github.com/larkox)

Start and join voice calls, video calls and use screen sharing with your team members via MS Teams Meetings.

Usage
-----

Once enabled, clicking the video icon in a Mattermost channel invites team members to join an MS Teams meeting, hosted using the credentials of the user who initiated the call.

![Screenshot](https://user-images.githubusercontent.com/177788/42196048-af54d2b8-7e30-11e8-80a0-5e160ae06f03.png)


## Configuration

### Step 1: Create Mattermost App in Azure

Sign into [portal.azure.com](https://portal.azure.com) using an admin Azure account.

#### Azure Active Directory

Navigate to [App Registrations](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps)

Click `New registration` at the top of the page.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76347903-be67f580-62dd-11ea-829e-236dd45865a8.png"/>

Then fill out the form with the following values:

- Name: `Mattermost MS Teams Meetings Plugin`
- Supported account types: Default value (Single tenant)
- Redirect URI: `https://(MM_SITE_URL)/plugins/com.mattermost.msteamsmeetings/oauth2/complete`

Replace `(MM_SITE_URL)` with your Mattermost server's Site URL. Then submit the form by clicking `Register`.

<img width="700" src="https://user-images.githubusercontent.com/6913320/76348298-55cd4880-62de-11ea-8e0e-4ace3a8f8fcb.png"/>

Navigate to `Certificates & secrets` in the left pane.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76348833-3d116280-62df-11ea-8b13-d39a0a2f2024.png"/>

Click `New client secret`. Then click `Add`, and copy the new secret on the bottom right corner of the screen. We'll use this value later in the Mattermost admin console.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76349025-9da09f80-62df-11ea-8c8f-0b39cad4597e.png"/>

Navigate to `API permissions` in the left pane.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76349582-a9d92c80-62e0-11ea-9414-5efd12c09b3f.png"/>

Click `Add a permission`, then click `Microsoft Graph` in the right pane.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76350226-c2961200-62e1-11ea-9080-19a9b75c2aee.png"/>

Click `Delegated permissions`, and scroll down to select the following permissions:
- OnlineMeetings.ReadWrite 

<img width="500" src="https://user-images.githubusercontent.com/6913320/76350551-5a93fb80-62e2-11ea-8eb3-812735691af9.png"/>

Submit the form by clicking `Add permissions` at the bottom.

<img width="500" src="https://user-images.githubusercontent.com/6913320/80412303-abb07c80-889b-11ea-9640-7c2f264c790f.png"/>

Click `Grant admin consent for...` to grant the permissions for the application. You're all set for configuration inside of Azure.

### Step 2: Configure Plugin Settings

Copy the `Client ID` and `Tenant ID` from the Azure portal

<img width="500" src="https://user-images.githubusercontent.com/6913320/76779336-9109c480-6781-11ea-8cde-4b79e5b2f3cd.png"/>

**`System Console` > `PLUGINS` > `Microsoft Teams Meetings`**

- `tenantID` - copy from Azure App
- `clientID` - copy from Azure App
- `Client Secret` - copy from Azure App (Generated in `Certificates & secrets`, earlier in these instructions)

## Development

This plugin contains both a server and web app portion.

Use `make dist` to build distributions of the plugin that you can upload to a Mattermost server for testing.

Use `make check-style` to check the style for the whole plugin.

### Server

Inside the `/server` directory, you will find the Go files that make up the server-side of the plugin. Within there, build the plugin like you would any other Go application.

### Web App

Inside the `/webapp` directory, you will find the JS and React files that make up the client-side of the plugin. Within there, modify files and components as necessary. Test your syntax by running `npm run build`.
