import addons, { mockChannel } from "@storybook/addons";
import React, { createContext, useContext, useState } from "react";
import useLocalStorage from "./useLocalStorage";
import useMediaQuery from "./useMediaQuery";
import { classNames } from "../utils/styles";
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
  localStorageTheme: string | null;
  theme: Theme;
  withFooterPadding: boolean;
  wrapperRef: React.Ref<null>;
  setTheme(theme: string): void;
  setWithFooterPadding(newValue: boolean): void;
  setWrapperRef(element: any): void;
}

const ThemeContext = createContext<IThemeContext | undefined>(undefined);

const ThemeProvider = ({ children, storyMode = false }) => {
  const [withFooterPadding, setWithFooterPadding] = useState(true);
  const [localStorageTheme, setLocalStorageTheme] =
    useLocalStorage("steampipe.ui.theme");
  const prefersDarkTheme = useMediaQuery("(prefers-color-scheme: dark)");
  const [wrapperRef, setWrapperRef] = useState(null);
  const doSetWrapperRef = (element) => setWrapperRef(() => element);

  // useEffect(() => {
  //   if (!themeConfig) {
  //     return;
  //   }
  //   // console.log(themeConfig);
  //   // console.log(themeConfig.theme.colors.foreground);
  //   let foregroundVar = themeConfig.theme.colors.foreground.replace("var(", "");
  //   foregroundVar = foregroundVar.replace(")", "");
  //   console.log(foregroundVar);
  //   console.log(document.documentElement.style.getPropertyValue(foregroundVar));
  // }, [themeConfig]);

  let theme;

  if (useDarkMode() && storyMode) {
    theme = Themes[ThemeNames.STEAMPIPE_DARK];
  } else if (storyMode) {
    theme = Themes[ThemeNames.STEAMPIPE_DEFAULT];
  } else if (
    localStorageTheme &&
    (localStorageTheme === ThemeNames.STEAMPIPE_DEFAULT ||
      localStorageTheme === ThemeNames.STEAMPIPE_DARK)
  ) {
    theme = Themes[localStorageTheme];
  } else if (prefersDarkTheme) {
    theme = Themes[ThemeNames.STEAMPIPE_DARK];
  } else {
    theme = Themes[ThemeNames.STEAMPIPE_DEFAULT];
  }

  return (
    <ThemeContext.Provider
      value={{
        localStorageTheme,
        theme,
        setTheme: setLocalStorageTheme,
        setWrapperRef: doSetWrapperRef,
        withFooterPadding,
        wrapperRef,
        setWithFooterPadding,
      }}
    >
      {children}
    </ThemeContext.Provider>
  );
};

const ThemeWrapper = ({ children }) => {
  const { setWrapperRef, theme } = useTheme();
  return (
    <div
      ref={setWrapperRef}
      className={`theme-${theme.name} bg-dashboard text-foreground`}
    >
      {children}
    </div>
  );
};

const FullHeightThemeWrapper = ({ children }) => {
  const { setWrapperRef, theme, withFooterPadding } = useTheme();
  return (
    <div
      ref={setWrapperRef}
      className={classNames(
        `min-h-screen flex flex-col theme-${theme.name} bg-dashboard text-foreground`,
        withFooterPadding ? "pb-8" : ""
      )}
    >
      {children}
    </div>
  );
};

const ModalThemeWrapper = ({ children }) => {
  const { setWrapperRef, theme } = useTheme();
  return (
    <div ref={setWrapperRef} className={`theme-${theme.name} text-foreground`}>
      {children}
    </div>
  );
};

const useTheme = () => {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error("useTheme must be used within a ThemeContext");
  }
  return context;
};

export {
  FullHeightThemeWrapper,
  ModalThemeWrapper,
  Themes,
  ThemeNames,
  ThemeProvider,
  ThemeWrapper,
  useTheme,
};
