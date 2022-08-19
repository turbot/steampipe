import { useDashboard } from "./useDashboard";
import { useEffect, useState } from "react";

const useChartThemeColors = () => {
  const {
    themeContext: { theme, wrapperRef },
  } = useDashboard();

  const getThemeColors = () => {
    // We need to get the theme CSS variable values - these are accessible on the theme root element and below in the tree
    const style = wrapperRef
      ? // @ts-ignore
        window.getComputedStyle(wrapperRef)
      : null;
    if (style) {
      const blackScale3 = style
        .getPropertyValue("--color-black-scale-3")
        .trim();
      const blackScale4 = style
        .getPropertyValue("--color-black-scale-4")
        .trim();
      const foreground = style.getPropertyValue("--color-foreground").trim();
      const foregroundLightest = style
        .getPropertyValue("--color-foreground-lightest")
        .trim();
      const alert = style.getPropertyValue("--color-alert").trim();
      const info = style.getPropertyValue("--color-info").trim();
      const ok = style.getPropertyValue("--color-ok").trim();
      return {
        blackScale3,
        blackScale4,
        foreground,
        foregroundLightest,
        alert,
        info,
        ok,
      };
    } else {
      return {};
    }
  };

  const [themeColors, setThemeColors] = useState(getThemeColors());

  useEffect(() => {
    setThemeColors(getThemeColors());
    // getThemeColors uses a ref that can sit outside the hook dependencies
  }, [theme.name, setThemeColors]);

  return themeColors;
};

export default useChartThemeColors;
