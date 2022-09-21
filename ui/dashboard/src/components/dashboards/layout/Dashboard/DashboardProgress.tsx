import { DashboardRunState } from "../../../../hooks/useDashboard";

interface DashboardProgressProps {
  state?: DashboardRunState;
  progress?: number;
}

const DashboardProgress = ({ state, progress }: DashboardProgressProps) => {
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
