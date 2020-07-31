import {Channel} from 'mattermost-redux/types/channels';

export interface PluginRegistry {
    registerChannelHeaderButtonAction(icon: React.ReactNode, callback: (channel: Channel) => void, text: string)
    registerPostTypeComponent(typeName: string, component: React.ElementType)
}
