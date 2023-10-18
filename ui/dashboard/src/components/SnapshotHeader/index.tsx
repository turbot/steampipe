import { classNames } from "../../utils/styles";
import { DashboardDataModeCLISnapshot } from "../../types";
import { useDashboard } from "../../hooks/useDashboard";

const SnapshotHeader = () => {
  const { dataMode, diff, snapshotFileName } = useDashboard();

  if (dataMode !== DashboardDataModeCLISnapshot) {
    return null;
  }

  return (
    <>
      <div
        className={classNames(
          "text-sm space-x-1 space-y-2 text-foreground-light",
        )}
      >
        <div>
          <span>Snapshot:</span>
          <span>{snapshotFileName}</span>
          {!!diff && diff.snapshotFileName && (
            <div>
              <span>Diff:</span>
              <span>{diff.snapshotFileName}</span>
            </div>
          )}
        </div>
      </div>
    </>
  );
};

export default SnapshotHeader;
