let heroIcons = {};
let materialSymbols = {};
import("../icons/heroIcons").then((m) => {
  heroIcons = m.icons;
});
import("../icons/materialSymbols").then((m) => {
  materialSymbols = m.icons;
});

const useDashboardIcons = () => ({
  heroIcons,
  materialSymbols,
});

export default useDashboardIcons;
