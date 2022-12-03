import { useMemo } from "react";

let heroIcons = {};
let materialSymbols = {};
import("../icons/heroIcons").then((m) => {
  console.log("Hero icons loaded");
  heroIcons = m.icons;
});
import("../icons/materialSymbols").then((m) => {
  console.log("Material symbols loaded");
  materialSymbols = m.icons;
});

const useDashboardIcons = () => {
  return useMemo(() => {
    console.log("Re-generating icons");
    return {
      heroIcons,
      materialSymbols,
    };
  }, [heroIcons, materialSymbols]);
};

export default useDashboardIcons;
