const path = require("path");
const webpack = require("webpack");
const { addBeforeLoader, loaderByName } = require("@craco/craco");
const { set } = require("lodash");

module.exports = {
  webpack: {
    configure: (webpackConfig) => {
      // console.log(webpackConfig.module.rules[1]);
      // const wasmExtensionRegExp = /\.wasm$/;
      // webpackConfig.resolve.extensions.push(".wasm");
      //
      // webpackConfig.module.rules.forEach((rule) => {
      //   (rule.oneOf || []).forEach((oneOf) => {
      //     if (oneOf.loader && oneOf.loader.indexOf("file-loader") >= 0) {
      //       oneOf.exclude.push(wasmExtensionRegExp);
      //     }
      //   });
      // });
      //
      // webpackConfig.module.rules.push({
      //   test: wasmExtensionRegExp,
      //   include: path.resolve(__dirname, "src"),
      //   use: [{ loader: require.resolve("wasm-loader"), options: {} }],
      // });
      // const wasmLoader = {
      //   test: wasmExtensionRegExp,
      //   include: path.resolve(__dirname, "src"),
      //   use: [{ loader: require.resolve("wasm-loader"), options: {} }],
      // };

      // addBeforeLoader(webpackConfig, loaderByName("file-loader"), wasmLoader);

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
