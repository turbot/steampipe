const set = require("lodash/set");
const webpack = require("webpack");

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
      webpackConfig = set(webpackConfig, "resolve.fallback.fs", false);
      webpackConfig = set(webpackConfig, "experiments.asyncWebAssembly", true);
      // webpackConfig = set(webpackConfig, "experiments.syncWebAssembly", true);
      return webpackConfig;
    },
    plugins: [
      new webpack.ProvidePlugin({
        Buffer: ["buffer", "Buffer"],
        process: "process/browser.js",
      }),
    ],
  },
};
