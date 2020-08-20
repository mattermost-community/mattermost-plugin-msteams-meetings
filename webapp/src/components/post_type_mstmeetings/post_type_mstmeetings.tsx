// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import React from 'react';

import {makeStyleFromTheme} from 'mattermost-redux/utils/theme_utils';
import {Post} from 'mattermost-redux/types/posts';
import {Theme} from 'mattermost-redux/types/preferences';

import {Svgs} from '../../constants';
import {formatDate} from '../../utils/date_utils';

type Props = {
    post: Post;
    compactDisplay?: boolean;
    isRHS?: boolean;
    useMilitaryTime?: boolean;
    theme: Theme;
    creatorName: string;
    currentChannelId: string;
    fromBot: boolean;
    actions: {
        startMeeting: (channelID: string, force: boolean) => void;
    };
}

export default class PostTypeMSTMeetings extends React.PureComponent<Props> {
    static defaultProps = {
        compactDisplay: false,
        isRHS: false,
    };

    render() {
        const style = getStyle(this.props.theme);
        const post = this.props.post;
        const props = post.props || {};

        let preText = '';
        let content: JSX.Element | undefined;
        let subtitle = '';
        if (props.meeting_status === 'STARTED') {
            preText = 'I have started a meeting';
            if (this.props.fromBot) {
                preText = `${this.props.creatorName} has started a meeting`;
            }
            content = (
                <a
                    className='btn btn-lg btn-primary'
                    style={style.button}
                    rel='noopener noreferrer'
                    target='_blank'
                    href={props.meeting_link}
                >
                    <i
                        style={style.buttonIcon}
                        dangerouslySetInnerHTML={{__html: Svgs.VIDEO_CAMERA_3}}
                    />
                    {'JOIN MEETING'}
                </a>
            );
        } else if (props.meeting_status === 'ENDED') {
            preText = 'I have ended the meeting';
            if (this.props.fromBot) {
                preText = `${this.props.creatorName} has ended the meeting`;
            }

            if (props.meeting_personal) {
                subtitle = 'Personal Meeting ID (PMI) : ' + props.meeting_id;
            } else {
                subtitle = 'Meeting ID : ' + props.meeting_id;
            }

            const startDate = new Date(post.create_at);
            const start = formatDate(startDate);
            const length = Math.ceil((new Date(post.update_at).getTime() - startDate.getTime()) / 1000 / 60);

            content = (
                <div>
                    <h2 style={style.summary}>
                        {'Meeting Summary'}
                    </h2>
                    <span style={style.summaryItem}>{'Date: ' + start}</span>
                    <br/>
                    <span style={style.summaryItem}>{'Meeting Length: ' + length + ' minute(s)'}</span>
                </div>
            );
        } else if (props.meeting_status === 'RECENTLY_CREATED') {
            preText = `${this.props.creatorName} already created a MS Teams Meeting recently`;

            subtitle = 'Would you like to join, or create your own meeting?';
            content = (
                <div>
                    <div>
                        <a
                            className='btn btn-lg btn-primary'
                            style={style.button}
                            rel='noopener noreferrer'
                            onClick={() => this.props.actions.startMeeting(this.props.currentChannelId, true)}
                        >
                            {'CREATE NEW MEETING'}
                        </a>
                    </div>
                    <div>
                        <a
                            className='btn btn-lg btn-primary'
                            style={style.button}
                            rel='noopener noreferrer'
                            target='_blank'
                            href={props.meeting_link}
                        >
                            <i
                                style={style.buttonIcon}
                                dangerouslySetInnerHTML={{__html: Svgs.VIDEO_CAMERA_3}}
                            />
                            {'JOIN EXISTING MEETING'}
                        </a>
                    </div>
                </div>
            );
        }

        let title = 'MS Teams Meeting';
        if (props.meeting_topic) {
            title = props.meeting_topic;
        }

        return (
            <div className='attachment attachment--pretext'>
                <div className='attachment__thumb-pretext'>
                    {preText}
                </div>
                <div className='attachment__content'>
                    <div className='clearfix attachment__container'>
                        <h5
                            className='mt-1'
                            style={style.title}
                        >
                            {title}
                        </h5>
                        {subtitle}
                        <div>
                            <div style={style.body}>
                                {content}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

const getStyle = makeStyleFromTheme((theme) => {
    return {
        body: {
            overflowX: 'auto',
            overflowY: 'hidden',
            paddingRight: '5px',
            width: '100%',
        },
        title: {
            fontWeight: '600',
        },
        button: {
            fontFamily: 'Open Sans',
            fontSize: '12px',
            fontWeight: 'bold',
            letterSpacing: '1px',
            lineHeight: '19px',
            marginTop: '12px',
            borderRadius: '4px',
            color: theme.buttonColor,
        },
        buttonIcon: {
            paddingRight: '8px',
            fill: theme.buttonColor,
        },
        summary: {
            fontFamily: 'Open Sans',
            fontSize: '14px',
            fontWeight: '600',
            lineHeight: '26px',
            margin: '0',
            padding: '14px 0 0 0',
        },
        summaryItem: {
            fontFamily: 'Open Sans',
            fontSize: '14px',
            lineHeight: '26px',
        },
    };
});
