// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import React from 'react';
import {FormattedMessage} from 'react-intl';

export default class Icon extends React.PureComponent {
    render() {
        return (
            <FormattedMessage
                id='msteamsmeetings.camera.ariaLabel'
                defaultMessage='camera icon'
            >
                {(ariaLabel) => (
                    <span
                        aria-label={ariaLabel}
                    >
                        <i className='icon icon-brand-zoom'/>
                    </span>
                )}
            </FormattedMessage>
        );
    }
}
