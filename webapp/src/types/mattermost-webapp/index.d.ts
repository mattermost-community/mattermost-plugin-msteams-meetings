import {Channel} from 'mattermost-redux/types/channels';

export interface PluginRegistry {
    registerChannelHeaderButtonAction(icon: React.ReactNode, callback: (channel: Channel) => void, dropdownTown: string, tooltipText: string)
    registerPostTypeComponent(typeName: string, component: React.ElementType)
    registerAppBarComponent(iconURL: string, action: (channel: Channel, member: ChannelMembership) => void, tooltipText: React.ReactNode)
}
