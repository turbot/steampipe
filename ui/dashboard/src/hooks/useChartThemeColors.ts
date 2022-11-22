import { useDashboard } from "./useDashboard";
import { useCallback, useEffect, useState } from "react";

const useChartThemeColors = () => {
  const {
    themeContext: { theme, wrapperRef },
  } = useDashboard();

  const getThemeColors = useCallback(() => {
    // We need to get the theme CSS variable values - these are accessible on the theme root element and below in the tree
    const style = wrapperRef
      ? // @ts-ignore
        window.getComputedStyle(wrapperRef)
      : null;
    if (style) {
      const dashboard = style.getPropertyValue("--color-dashboard").trim();
      const dashboardPanel = style
        .getPropertyValue("--color-dashboard-panel")
        .trim();
      const blackScale3 = style
        .getPropertyValue("--color-black-scale-3")
        .trim();
      const blackScale4 = style
        .getPropertyValue("--color-black-scale-4")
        .trim();
      const foreground = style.getPropertyValue("--color-foreground").trim();
      const foregroundLight = style
        .getPropertyValue("--color-foreground-light")
        .trim();
      const foregroundLighter = style
        .getPropertyValue("--color-foreground-lighter")
        .trim();
      const foregroundLightest = style
        .getPropertyValue("--color-foreground-lightest")
        .trim();
      const alert = style.getPropertyValue("--color-alert").trim();
      const info = style.getPropertyValue("--color-info").trim();
      const ok = style.getPropertyValue("--color-ok").trim();
      return {
        dashboard,
        dashboardPanel,
        blackScale3,
        blackScale4,
        foreground,
        foregroundLight,
        foregroundLighter,
        foregroundLightest,
        alert,
        info,
        ok,
      };
    } else {
      return {};
    }
  }, [wrapperRef]);

  const [themeColors, setThemeColors] = useState(getThemeColors());
  const [random, setRandom] = useState<number | null>(null);

  useEffect(() => {
    setThemeColors(getThemeColors());
    // getThemeColors uses a ref that can sit outside the hook dependencies
  }, [getThemeColors, random, theme.name, setThemeColors]);

  useEffect(() => {
    setRandom(Math.random() * Math.random());
  }, [theme]);

  return themeColors;
};

export default useChartThemeColors;
