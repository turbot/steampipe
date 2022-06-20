import { DashboardRunState } from "../../../../hooks/useDashboard";

interface DashboardProgressProps {
  state?: DashboardRunState;
  progress?: number;
}

const DashboardProgress = ({ state, progress }: DashboardProgressProps) => {
  return (
    <div className="w-full h-[4px] bg-dashboard">
      {state === "running" ? (
        <div
          className="h-full dashboard-loading-animate"
          style={{ width: `${progress}%` }}
        />
      ) : null}
    </div>
  );
};

export default DashboardProgress;
