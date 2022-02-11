import { AvailableDashboard, useDashboard } from "../../hooks/useDashboard";
import { Link } from "react-router-dom";
import { useEffect, useState } from "react";

const DashboardList = () => {
  const {
    availableDashboardsLoaded,
    metadataLoaded,
    metadata,
    dashboards,
    selectedDashboard,
  } = useDashboard();
  const [dashboardsForCurrentMod, setDashboardsForCurrentMod] = useState<
    AvailableDashboard[]
  >([]);
  const [dashboardsForOtherMod, setDashboardsForOtherMod] = useState<
    AvailableDashboard[]
  >([]);

  useEffect(() => {}, []);

  if (selectedDashboard) {
    return null;
  }

  // console.log(metadata);
  // console.log(dashboards);

  return (
    <div className="w-full p-4">
      {!availableDashboardsLoaded && !metadataLoaded && <></>}
      {availableDashboardsLoaded && metadataLoaded && (
        <>
          <h2 className="text-2xl font-medium">
            Welcome to the Steampipe Dashboard UI
          </h2>
          {dashboards.length === 0 ? (
            <p className="py-2 italic">
              No dashboards defined. Please define at least one dashboard in the
              mod.
            </p>
          ) : null}
          {dashboards.length > 0 ? (
            <p className="py-2">
              Please select a dashboard from the list below:
            </p>
          ) : null}
          <ul className="list-none list-inside">
            {dashboards.map((availableDashboard) => (
              <li key={availableDashboard.full_name} className="pl-4">
                <Link
                  className="link-highlight"
                  to={`/${availableDashboard.full_name}`}
                >
                  {availableDashboard.title || availableDashboard.full_name}
                </Link>
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
};

export default DashboardList;
