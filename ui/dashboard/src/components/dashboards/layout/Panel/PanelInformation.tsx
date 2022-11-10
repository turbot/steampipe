import { classNames } from "../../../../utils/styles";
import { ThemeNames } from "../../../../hooks/useTheme";
import { useDashboard } from "../../../../hooks/useDashboard";
import { usePanel } from "../../../../hooks/usePanel";

const PanelInformation = () => {
  const {
    themeContext: { theme },
  } = useDashboard();
  const { showPanelInformation, panelInformation } = usePanel();

  if (!showPanelInformation) {
    return null;
  }

  return (
    <div
      className={classNames(
        "absolute h-full overflow-y-scroll z-50 top-0 right-0 p-3 max-w-sm bg-dashboard-panel border-l text-sm",
        theme.name === ThemeNames.STEAMPIPE_DARK
          ? "border-table-divide"
          : "border-background"
      )}
    >
      {panelInformation}
    </div>
  );
};

export default PanelInformation;
