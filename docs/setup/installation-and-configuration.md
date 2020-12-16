## Installation

The Mattermost MS Teams Meetings plugin is provided in the Mattermost Plugin Marketplace. Once enabled, clicking the video icon in a Mattermost channel invites team members to join an MS Teams meeting, hosted using the credentials of the user who initiated the call.

### Requirements

* Mattermost Server v5.26+ is required.
* Mattermost Enterprise Edition E20 is required.

### Marketplace Installation

1. Go to **Main Menu > Plugin Marketplace** in Mattermost.
2. Search for "MS Teams" or manually find the plugin from the list and click **Install**.
3. After the plugin is downloaded and installed, select **Configure**.

### Manual Installation

If your server doesn't have access to the internet, you can download the latest [plugin binary release](https://github.com/mattermost/mattermost-plugin-msteams-meetings/releases) and upload it to your server via **System Console > Plugin Management**. The releases on this page are the same versions available on the Plugin Marketplace.

## Configuration

### Step 1: Create a Mattermost App in Azure

1. Sign in to [the Azure portal](https://portal.azure.com) using an admin Azure account.
2. Navigate to [App Registrations](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps).
3. Select **New registration** at the top of the page.

<img width="300" src="https://user-images.githubusercontent.com/6913320/76347903-be67f580-62dd-11ea-829e-236dd45865a8.png"/>

4. Fill out the form with the following values:

- Name: **Mattermost MS Teams Meetings Plugin**
- Supported account types: **Default value (Single tenant)**
- Redirect URI: **https://(MM_SITE_URL)/plugins/com.mattermost.msteamsmeetings/oauth2/complete**. Replace `(MM_SITE_URL)` with your Mattermost server's Site URL. Select **Register** to submit the form.

<img width="500 src="https://user-images.githubusercontent.com/6913320/76348298-55cd4880-62de-11ea-8e0e-4ace3a8f8fcb.png"/>

5. Navigate to **Certificates & secrets** in the left pane.

<img width="300" src="https://user-images.githubusercontent.com/6913320/76348833-3d116280-62df-11ea-8b13-d39a0a2f2024.png"/>

6. Select **New client secret > Add**, then copy the new secret in the bottom right corner of the screen. We'll use this value later in the Mattermost System Console.

<img width="300" src="https://user-images.githubusercontent.com/6913320/76349025-9da09f80-62df-11ea-8c8f-0b39cad4597e.png"/>

7. Navigate to **API permissions** in the left pane.

<img width="300" src="https://user-images.githubusercontent.com/6913320/76349582-a9d92c80-62e0-11ea-9414-5efd12c09b3f.png"/>

8. Select **Add a permission** and choose **Microsoft Graph** in the right pane.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76350226-c2961200-62e1-11ea-9080-19a9b75c2aee.png"/>

9. Select **Delegated permissions**, and scroll down to select the `OnlineMeetings.ReadWrite` permissions.

<img width="300" src="https://user-images.githubusercontent.com/6913320/76350551-5a93fb80-62e2-11ea-8eb3-812735691af9.png"/>

10. Select **Add permissions** to submit the form.

<img width="300" src="https://user-images.githubusercontent.com/6913320/80412303-abb07c80-889b-11ea-9640-7c2f264c790f.png"/>

11. Select **Grant admin consent for...** to grant the permissions for the application.

You're all set for configuration inside of the Azure portal.

### Step 2: Configure Plugin Settings

1. Copy the **Client ID** and **Tenant ID** from the Azure portal.

<img width="300" src="https://user-images.githubusercontent.com/6913320/76779336-9109c480-6781-11ea-8cde-4b79e5b2f3cd.png"/>

2. Go to **System Console > Plugins > MS Teams Meetings**.
3. Enter the following values in the fields provided:

- `tenantID` - copy from the Azure portal
- `clientID` - copy from the Azure portal
- `Client Secret` - copy from the Azure portal (generated in **Certificates & secrets** earlier in these instructions)

4. Choose **Save** to apply the configuration.

## Onboarding users

When you’ve tested the plugin and confirmed it’s working, notify your team so they can get started. Copy and paste the text below, edit it to suit your requirements, and send it out.

> Hi team,

> The MS Teams Meetings plugin has been configured so you can use it for calls from within Mattermost. To get started, run the `/mstmeetings connect` slash command from any channel within Mattermost. Visit the documentation for more information.
