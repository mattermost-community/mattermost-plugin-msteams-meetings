# Enviroment

This plugin contains both a server and web app portion. Read our documentation about the [Developer Workflow](https://developers.mattermost.com/extend/plugins/developer-workflow/) and [Developer Setup](https://developers.mattermost.com/extend/plugins/developer-setup/) for more information about developing and extending plugins.

## Server

Inside the `/server` directory, you will find the Go files that make up the server-side of the plugin. Within that directory, build the plugin like you would any other Go application.

## Web App

Inside the `/webapp` directory, you will find the JavaScript files that make up the client-side of the plugin. Within that directory, modify files and components as necessary. Test your syntax by running `npm run build`.

## Deploying to a Mattermost local server

It's on the [Developer setup](https://developers.mattermost.com/integrate/plugins/developer-setup/#deploy-with-local-mode), but keep in mind It's necessary to enable PluginUploads.

Then you need to set these envvars to be able to upload the plugin:

```bash
export MM_SERVICESETTINGS_SITEURL=https://localhost:8065/   # Or other if needed
export MM_ADMIN_USERNAME=<MYUSERNAME>
export MM_ADMIN_PASSWORD=<MYPASSWORD>
```

Then run `make deploy` or `MM_DEBUG=1 make deploy` in case you want debugging from the root of the `mattermost-plugin-msteams-meetings` project.
