const webpack = require("webpack");
module.exports = {
    publicPath:"",

    configureWebpack:{
        plugins: [
            // new webpack.ProvidePlugin({
            //     $: 'jquery',
            //     jquery: 'jquery',
            //     'window.jQuery': 'jquery',
            //     jQuery: 'jquery'
            // }),
        ],
    }
};
