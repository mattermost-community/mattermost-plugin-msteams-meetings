// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import React from 'react';
import {IntlProvider, FormattedMessage} from 'react-intl';

export default class Icon extends React.PureComponent {
    render() {
        return (
            <IntlProvider locale='en'>
                <FormattedMessage
                    id='msteamsmeetings.camera.ariaLabel'
                    defaultMessage='camera icon'
                >
                    {(ariaLabel: string) => (
                        <span
                            aria-label={ariaLabel}
                        >
                            <i className='icon icon-brand-zoom'/>
                        </span>
                    )}
                </FormattedMessage>
            </IntlProvider>
        );
    }
}
