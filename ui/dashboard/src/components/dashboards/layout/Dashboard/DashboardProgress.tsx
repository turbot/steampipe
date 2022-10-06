import { DashboardDataModeLive } from "../../../../types";
import { useDashboard } from "../../../../hooks/useDashboard";

const DashboardProgress = () => {
  const { dataMode, progress, state } = useDashboard();

  // We only show a progress indicator in live mode
  if (dataMode !== DashboardDataModeLive) {
    return null;
  }

  return (
    <div className="w-full h-[4px] bg-dashboard print:hidden">
      {state === "ready" ? (
        <div
          className="h-full bg-black-scale-3"
          style={{ width: `${progress}%` }}
        />
      ) : null}
    </div>
  );
};

export default DashboardProgress;
