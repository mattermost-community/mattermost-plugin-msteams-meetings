// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import {connect} from 'react-redux';
import {ActionCreatorsMapObject, bindActionCreators, Dispatch} from 'redux';

import {getBool} from 'mattermost-redux/selectors/entities/preferences';
import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/common';
import {ActionResult} from 'mattermost-redux/types/actions';
import {Post} from 'mattermost-redux/types/posts';
import {GlobalState} from 'mattermost-redux/types/store';
import {Theme} from 'mattermost-redux/types/preferences';

import {warnAndConfirm, startMeeting} from '../../actions';

import PostTypeMSTMeetings from './post_type_mstmeetings';

type OwnProps = {
    post: Post;
    compactDisplay?: boolean;
    isRHS?: boolean;
    theme: Theme;
    currentChannelId: string;
}

type Actions = {
    startMeeting: (channelID: string, force: boolean) => ActionResult;
    warnAndConfirm: (channelID: string) => ActionResult;
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
        actions: bindActionCreators<ActionCreatorsMapObject, Actions>({
            warnAndConfirm,
            startMeeting,
        }, dispatch),
    };
}

export default connect(mapStateToProps, mapDispatchToProps)(PostTypeMSTMeetings);
