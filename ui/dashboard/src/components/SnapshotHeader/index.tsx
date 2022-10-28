import { classNames } from "../../utils/styles";
import { DashboardDataModeCLISnapshot } from "../../types";
import { ThemeNames, useTheme } from "../../hooks/useTheme";
import { useDashboard } from "../../hooks/useDashboard";

const SnapshotBanner = () => {
  const { dataMode, snapshotFileName } = useDashboard();
  const { theme } = useTheme();

  if (dataMode !== DashboardDataModeCLISnapshot) {
    return null;
  }

  return (
    <div
      className={classNames(
        "w-full bg-dashboard border-b p-4 text-sm space-x-1 text-foreground-light",
        theme.name === ThemeNames.STEAMPIPE_DARK
          ? "border-table-divide"
          : "border-background"
      )}
    >
      <span>Snapshot:</span>
      <span>{snapshotFileName}</span>
    </div>
  );
};

export default SnapshotBanner;
