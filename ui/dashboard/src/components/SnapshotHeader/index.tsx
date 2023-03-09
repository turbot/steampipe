import { classNames } from "../../utils/styles";
import { DashboardDataModeCLISnapshot } from "../../types";
import { useDashboard } from "../../hooks/useDashboard";

const SnapshotBanner = () => {
  const { dataMode, snapshotFileName } = useDashboard();

  if (dataMode !== DashboardDataModeCLISnapshot) {
    return null;
  }

  return (
    <div
      className={classNames(
        "w-full bg-dashboard border-b border-divide p-4 text-sm space-x-1 text-foreground-light"
      )}
    >
      <span>Snapshot:</span>
      <span>{snapshotFileName}</span>
    </div>
  );
};

export default SnapshotBanner;
