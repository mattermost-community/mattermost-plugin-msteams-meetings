var path = require('path');

module.exports = {
    entry: [
        './src/index.tsx',
    ],
    resolve: {
        modules: [
            'src',
            'node_modules',
            path.resolve(__dirname),
        ],
        extensions: ['*', '.js', '.jsx', '.ts', '.tsx'],
    },
    module: {
        rules: [
            {
                test: /\.(js|jsx|ts|tsx)$/,
                exclude: /node_modules/,
                use: {
                    loader: 'babel-loader',
                    options: {
                        plugins: [
                            '@babel/plugin-proposal-class-properties',
                            '@babel/plugin-syntax-dynamic-import',
                            '@babel/proposal-object-rest-spread',
                            'babel-plugin-typescript-to-proptypes',
                        ],
                        presets: [
                            ['@babel/preset-env', {
                                targets: {
                                    chrome: 66,
                                    firefox: 60,
                                    edge: 42,
                                    safari: 12,
                                },
                                corejs: 3,
                                modules: false,
                                debug: false,
                                useBuiltIns: 'usage',
                                shippedProposals: true,
                            }],
                            ['@babel/preset-react', {
                                useBuiltIns: true,
                            }],
                            ['@babel/typescript', {
                                allExtensions: true,
                                isTSX: true,
                            }],
                        ],
                    },
                },
            },
        ],
    },
    externals: {
        react: 'React',
        redux: 'Redux',
        'react-redux': 'ReactRedux',
        'prop-types': 'PropTypes',
        'react-bootstrap': 'ReactBootstrap',
    },
    output: {
        path: path.join(__dirname, '/dist'),
        publicPath: '/',
        filename: 'main.js',
    },
};
