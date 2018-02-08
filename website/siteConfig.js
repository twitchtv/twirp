/**
 * Copyright (c) 2017-present, Facebook, Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

/* List of projects/orgs using your project for the users page */
const users = [
    {
        caption: 'Twitch',
        image: '/test-site/img/twitch.png',
        infoLink: 'https://twitch.tv',
        pinned: true,
    },
];

const siteConfig = {
    title: 'Twirp' /* title for your website */,
    tagline: 'Simple RPC framework powered by protobuf',
    url: 'https://twitchtv.github.io' /* your website url */,
    baseUrl: '/twirp/' /* base url for your project */,
    organizationName: 'twitchtv',
    projectName: 'twirp',
    headerLinks: [
        {doc: 'intro', label: 'Docs'},
        {doc: 'spec_v5', label: 'Spec'},
    ],
    users,
    /* colors for website */
    colors: {
        primaryColor: '#6441a5',
        secondaryColor: '#f1f1f1',
    },
    // This copyright info is used in /core/Footer.js and blog rss/atom feeds.
    copyright:
    'Copyright © ' +
        new Date().getFullYear() +
        ' Twitch Interactive, Inc.',
    // organizationName: 'deltice', // or set an env variable ORGANIZATION_NAME
    // projectName: 'test-site', // or set an env variable PROJECT_NAME
    highlight: {
        // Highlight.js theme to use for syntax highlighting in code blocks
        theme: 'tomorrow',
    },
    scripts: ['https://buttons.github.io/buttons.js'],
    // You may provide arbitrary config keys to be used as needed by your template.
    repoUrl: 'https://github.com/twitchtv/twirp',
};

module.exports = siteConfig;
