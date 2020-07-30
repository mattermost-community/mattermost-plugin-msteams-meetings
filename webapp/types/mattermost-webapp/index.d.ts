declare module 'mattermost-webapp/plugins/registry';
interface PluginRegistry {
    registerChannelHeaderButtonAction(icon: JSX.Element, callback: (channelID: any) => void, text: string)
    registerPostTypeComponent(typeName: string, component: any)
}