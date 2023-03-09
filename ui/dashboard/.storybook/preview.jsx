import StoryWrapper from "./StoryWrapper";
import { ThemeProvider } from "../src/hooks/useStorybookTheme";
import { themes } from "@storybook/theming";
import { withRouter } from "storybook-addon-react-router-v6";
import "../src/styles/index.css";

const viewports = {
  xs: {
    name: "Extra Small",
    styles: {
      width: "400px",
      height: "500px",
    },
    type: "mobile",
  },
  sm: {
    name: "Small",
    styles: {
      width: "640px",
      height: "800px",
    },
    type: "mobile",
  },
  md: {
    name: "Medium",
    styles: {
      width: "768px",
      height: "800px",
    },
    type: "tablet",
  },
  lg: {
    name: "Large",
    styles: {
      width: "1024px",
      height: "800px",
    },
    type: "desktop",
  },
  xl: {
    name: "Extra Large",
    styles: {
      width: "1280px",
      height: "900px",
    },
    type: "desktop",
  },
  "2xl": {
    name: "Extra Extra Large",
    styles: {
      width: "1536px",
      height: "1000px",
    },
    type: "desktop",
  },
};

export const decorators = [
  withRouter,
  (Story) => (
    <ThemeProvider>
      <StoryWrapper>
        <Story />
      </StoryWrapper>
    </ThemeProvider>
  ),
];

export const parameters = {
  actions: { argTypesRegex: "^on[A-Z].*" },
  darkMode: {
    // Override the default dark theme
    dark: {
      ...themes.dark,
      appContentBg: "#181717",
    },
    // Override the default light theme
    light: { ...themes.normal },
  },
  viewport: {
    viewports,
  },
};
