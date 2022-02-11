import ThemeToggle from "../ThemeToggle";
import { useNavigate, useParams } from "react-router-dom";
import { useDashboard } from "../../hooks/useDashboard";

const DashboardSelector = () => {
  const navigate = useNavigate();
  const { dashboards } = useDashboard();
  const { dashboardName } = useParams();
  return (
    <div className="flex justify-between items-center">
      <div className="flex items-center space-x-2">
        <select
          className="block w-full pl-3 pr-10 py-2 text-base bg-background border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          id="dashboard"
          name="dashboard"
          onChange={(e) => {
            const dashboardId = e.target.value;
            navigate(`/${dashboardId}`);
          }}
          value={dashboardName || ""}
        >
          <option value="">Choose a dashboard...</option>
          {dashboards.map((dashboard) => (
            <option key={dashboard.full_name} value={dashboard.full_name}>
              {dashboard.title || dashboard.full_name}
            </option>
          ))}
        </select>
      </div>
      <ThemeToggle />
    </div>
  );
};

export default DashboardSelector;
