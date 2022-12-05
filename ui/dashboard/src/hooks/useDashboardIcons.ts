import { useMemo } from "react";

let heroIcons = {};
let materialSymbols = {};
import("../icons/heroIcons").then((m) => {
  heroIcons = m.icons;
});
import("../icons/materialSymbols").then((m) => {
  materialSymbols = m.icons;
});

const useDashboardIcons = () => {
  return useMemo(() => {
    return {
      heroIcons,
      materialSymbols,
    };
  }, [heroIcons, materialSymbols]);
};

export default useDashboardIcons;
