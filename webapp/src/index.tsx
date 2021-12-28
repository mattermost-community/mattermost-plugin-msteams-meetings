// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import React from 'react';
import {Store, Action} from 'redux';

import {Channel} from 'mattermost-redux/types/channels';
import {GlobalState} from 'mattermost-redux/types/store';
import {getConfig} from 'mattermost-redux/selectors/entities/general';

import {id as pluginId} from './manifest';
import Icon from './components/icon';
import PostTypeMSTMeetings from './components/post_type_mstmeetings';
import {startMeeting} from './actions';
import Client from './client';
import {getPluginURL, getServerRoute} from './selectors';

// eslint-disable-next-line import/no-unresolved
import {PluginRegistry} from './types/mattermost-webapp';

class Plugin {
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        let creatingMeeting = false;
        registry.registerChannelHeaderButtonAction(
            <Icon/>,
            async (channel: Channel) => {
                if (!creatingMeeting) {
                    creatingMeeting = true;
                    await startMeeting(channel.id)(store.dispatch, store.getState);
                    creatingMeeting = false;
                }
            },
            'Start MS Teams Meeting',
            'Start MS Teams Meeting',
        );
        registry.registerPostTypeComponent('custom_mstmeetings', PostTypeMSTMeetings);
        Client.setServerRoute(getServerRoute(store.getState()));

        if (registry.registerAppBarComponent) {
            const appBarIconPath = '/public/app-bar-icon.png';
            const pluginURL = getPluginURL(store.getState());
            const iconURL = pluginURL + appBarIconPath;

            registry.registerAppBarComponent(
                iconURL,
                async (channel: Channel) => {
                    if (!creatingMeeting) {
                        creatingMeeting = true;
                        await startMeeting(channel.id)(store.dispatch, store.getState);
                        creatingMeeting = false;
                    }
                },
                'Start MS Teams Meeting',
            );
        }
    }
}

declare global {
    interface Window {
        registerPlugin(id: string, plugin: Plugin): void
    }
}

window.registerPlugin(pluginId, new Plugin());
