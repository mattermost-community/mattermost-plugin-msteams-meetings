// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import {connect} from 'react-redux';
import {ActionCreatorsMapObject, bindActionCreators, Dispatch} from 'redux';

import {getBool} from 'mattermost-redux/selectors/entities/preferences';
import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/common';
import {Post} from 'mattermost-redux/types/posts';
import {ActionFunc} from 'mattermost-redux/types/actions';
import {GlobalState} from 'mattermost-redux/types/store';

import {startMeeting} from '../../actions';

import PostTypeMSTMeetings from './post_type_mstmeetings';

type OwnProps = {

    /*
    * The post to render the message for.
    */
    post: Post;

    /**
     * Set to render post body compactly.
     */
    compactDisplay?: boolean;

    /**
     * Flags if the post_message_view is for the RHS (Reply).
     */
    isRHS?: boolean;

    /*
        * Logged in user's theme.
        */
    theme: any,

    /*
        * Current Channel Id.
        */
    currentChannelId: string,
}

type Actions = {
    startMeeting: (channelID: string, force: boolean) => void;
}

function mapStateToProps(state: GlobalState, ownProps: OwnProps) {
    return {
        ...ownProps,
        fromBot: ownProps.post.props.from_bot,
        creatorName: ownProps.post.props.meeting_creator_username || 'Someone',
        useMilitaryTime: getBool(state, 'display_settings', 'use_military_time', false),
        currentChannelId: getCurrentChannelId(state),
    };
}

function mapDispatchToProps(dispatch: Dispatch) {
    return {
        actions: bindActionCreators<ActionCreatorsMapObject<ActionFunc>, Actions>({
            startMeeting,
        }, dispatch),
    };
}

export default connect(mapStateToProps, mapDispatchToProps)(PostTypeMSTMeetings);
