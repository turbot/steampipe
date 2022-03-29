const set = require("lodash/set");

module.exports = {
  stories: ["../src/**/*.stories.mdx", "../src/**/*.stories.@(js|jsx|ts|tsx)"],
  addons: [
    "@storybook/addon-links",
    "@storybook/addon-essentials",
    "@storybook/preset-create-react-app",
    "storybook-dark-mode",
  ],
  core: {
    builder: "webpack5",
  },
  webpackFinal: async (config) => {
    config = set(config, "resolve.fallback.fs", false);
    return config;
  },
};
