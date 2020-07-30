// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import React from 'react';
import {Store, Action} from 'redux';
import {PluginRegistry} from 'mattermost-webapp/plugins/registry';
import {Channel} from 'mattermost-redux/types/channels';
import {GlobalState} from 'mattermost-redux/types/store';
import {getConfig} from 'mattermost-redux/selectors/entities/general';

import {id as pluginId} from './manifest';

import Icon from './components/icon';
import PostTypeMSTMeetings from './components/post_type_mstmeetings';
import {startMeeting} from './actions';
import Client from './client';

declare global {
    interface Window {
        registerPlugin(id: string, plugin: Plugin): void
    }
}

class Plugin {
    // eslint-disable-next-line no-unused-vars
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        registry.registerChannelHeaderButtonAction(
            <Icon/>,
            (channel: Channel) => {
                startMeeting(channel.id)(store.dispatch, store.getState);
            },
            'Start MS Teams Meeting',
        );
        registry.registerPostTypeComponent('custom_mstmeetings', PostTypeMSTMeetings);
        Client.setServerRoute(getServerRoute(store.getState()));
    }
}

window.registerPlugin(pluginId, new Plugin());

const getServerRoute = (state: GlobalState) => {
    const config = getConfig(state);

    let basePath = '';
    if (config && config.SiteURL) {
        basePath = new URL(config.SiteURL).pathname;

        if (basePath && basePath[basePath.length - 1] === '/') {
            basePath = basePath.substr(0, basePath.length - 1);
        }
    }

    return basePath;
};