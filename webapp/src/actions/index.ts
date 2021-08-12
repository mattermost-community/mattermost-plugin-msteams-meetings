// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import {Dispatch} from 'redux';

import {PostTypes} from 'mattermost-redux/action_types';
import {GetStateFunc} from 'mattermost-redux/types/actions';

import Client from '../client';

export function startMeeting(channelId: string, force = false) {
    return async (dispatch: Dispatch, getState: GetStateFunc) => {
        try {
            const startFunction = force ? Client.forceStartMeeting : Client.startMeeting;
            const meetingURL = await startFunction(channelId, true);
            if (meetingURL) {
                window.open(meetingURL);
            }

            return {data: true};
        } catch (error) {
            let m : string;
            if (error.message && error.message[0] === '{') {
                const e = JSON.parse(error.message);

                // Error is from MS API
                if (e?.error?.message) {
                    m = '\nMSTMeeting error: ' + e.error.message;
                } else {
                    m = e;
                }
            } else {
                m = error.message;
            }

            const post = {
                id: 'mstMeetingsPlugin' + Date.now(),
                create_at: Date.now(),
                update_at: 0,
                edit_at: 0,
                delete_at: 0,
                is_pinned: false,
                user_id: getState().entities.users.currentUserId,
                channel_id: channelId,
                root_id: '',
                parent_id: '',
                original_id: '',
                message: m,
                type: 'system_ephemeral',
                props: {},
                hashtags: '',
                pending_post_id: '',
            };

            dispatch({
                type: PostTypes.RECEIVED_NEW_POST,
                data: post,
                channelId,
            });

            return {error};
        }
    };
}

export function warnAndConfirm(channelId: string) {
    return async (dispatch: Dispatch, getState: GetStateFunc) => {
        try {
            const result = await Client.warnAndConfirmMeeting(channelId);
            return {result};
        } catch (error) {
            let m : string;
            if (error.message && error.message[0] === '{') {
                const e = JSON.parse(error.message);

                // Error is from MS API
                if (e?.error?.message) {
                    m = '\nMSTMeeting error: ' + e.error.message;
                } else {
                    m = e;
                }
            } else {
                m = error.message;
            }

            const post = {
                id: 'mstMeetingsPlugin' + Date.now(),
                create_at: Date.now(),
                update_at: 0,
                edit_at: 0,
                delete_at: 0,
                is_pinned: false,
                user_id: getState().entities.users.currentUserId,
                channel_id: channelId,
                root_id: '',
                parent_id: '',
                original_id: '',
                message: m,
                type: 'system_ephemeral',
                props: {},
                hashtags: '',
                pending_post_id: '',
            };

            dispatch({
                type: PostTypes.RECEIVED_NEW_POST,
                data: post,
                channelId,
            });

            return {error};
        }
    };
}
