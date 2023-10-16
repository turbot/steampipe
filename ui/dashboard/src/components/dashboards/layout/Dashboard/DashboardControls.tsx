import CheckGroupingConfig from "../../check/CheckGroupingConfig";
import SnapshotHeader from "../../../SnapshotHeader";
import { DashboardDataModeCLISnapshot } from "../../../../types";
import { useDashboard } from "../../../../hooks/useDashboard";

const DashboardControls = () => {
  const { dataMode, dashboard } = useDashboard();

  const isBenchmark =
    dashboard?.children && dashboard.children[0].panel_type === "benchmark";

  return (
    <div className="grid p-4 gap-4 grid-cols-3 bg-dashboard-panel print:hidden">
      {dataMode === DashboardDataModeCLISnapshot && <SnapshotHeader />}
      {isBenchmark && (
        <div className="col-span-3 space-y-4">
          <CheckGroupingConfig />
        </div>
      )}
    </div>
  );
};

export default DashboardControls;
