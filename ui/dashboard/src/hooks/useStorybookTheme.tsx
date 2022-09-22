import React, { createContext, useContext, useState } from "react";
import addons, { mockChannel } from "@storybook/addons";
import { useDarkMode } from "storybook-dark-mode";

if (!addons.hasChannel()) {
  addons.setChannel(mockChannel());
}

export interface Theme {
  name: string;
  label: string;
}

interface IThemes {
  [key: string]: Theme;
}

const ThemeNames = {
  STEAMPIPE_DEFAULT: "steampipe-default",
  STEAMPIPE_DARK: "steampipe-dark",
};

const Themes: IThemes = {
  [ThemeNames.STEAMPIPE_DEFAULT]: {
    label: "Light",
    name: ThemeNames.STEAMPIPE_DEFAULT,
  },
  [ThemeNames.STEAMPIPE_DARK]: {
    label: "Dark",
    name: ThemeNames.STEAMPIPE_DARK,
  },
};

interface IThemeContext {
  theme: Theme;
  setWithFooterPadding(newValue: boolean): void;
  setWrapperRef(element: any): void;
  withFooterPadding: boolean;
  wrapperRef: React.Ref<null>;
}

const ThemeContext = createContext<IThemeContext | undefined>(undefined);

const ThemeProvider = ({ children }) => {
  const [withFooterPadding, setWithFooterPadding] = useState(true);
  const [wrapperRef, setWrapperRef] = useState(null);
  const doSetWrapperRef = (element) => setWrapperRef(() => element);

  let theme;

  if (useDarkMode()) {
    theme = Themes[ThemeNames.STEAMPIPE_DARK];
  } else {
    theme = Themes[ThemeNames.STEAMPIPE_DEFAULT];
  }

  return (
    <ThemeContext.Provider
      value={{
        theme,
        setWithFooterPadding,
        setWrapperRef: doSetWrapperRef,
        withFooterPadding,
        wrapperRef,
      }}
    >
      {children}
    </ThemeContext.Provider>
  );
};

const ThemeWrapper = ({ children }) => {
  const { setWrapperRef, theme } = useStorybookTheme();
  return (
    <div
      ref={setWrapperRef}
      className={`theme-${theme.name} bg-dashboard print:bg-white print:theme-steampipe-default text-foreground print:text-black`}
    >
      {children}
    </div>
  );
};

const useStorybookTheme = () => {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error("useTheme must be used within a ThemeContext");
  }
  return context;
};

export { Themes, ThemeNames, ThemeProvider, ThemeWrapper, useStorybookTheme };
