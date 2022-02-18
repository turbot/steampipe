const webpack = require("webpack");
const { set } = require("lodash");

module.exports = {
  webpack: {
    configure: (webpackConfig) => {
      const scopePluginIndex = webpackConfig.resolve.plugins.findIndex(
        ({ constructor }) =>
          constructor && constructor.name === "ModuleScopePlugin"
      );

      webpackConfig.resolve.plugins.splice(scopePluginIndex, 1);

      webpackConfig = set(
        webpackConfig,
        "resolve.fallback.crypto",
        require.resolve("crypto-browserify")
      );
      webpackConfig = set(
        webpackConfig,
        "resolve.fallback.path",
        require.resolve("path-browserify")
      );
      webpackConfig = set(
        webpackConfig,
        "resolve.fallback.stream",
        require.resolve("stream-browserify")
      );
      return webpackConfig;
    },
    // configure: {
    //   resolve: {
    //     fallback: {
    //       crypto: require.resolve("crypto-browserify"),
    //       path: require.resolve("path-browserify"),
    //       stream: require.resolve("stream-browserify"),
    //     },
    //   },
    // },
    plugins: [
      new webpack.ProvidePlugin({
        Buffer: ["buffer", "Buffer"],
        process: "process/browser.js",
      }),
    ],
  },
};
