import LoadingIndicator from "../dashboards/LoadingIndicator";
import { AvailableDashboard, useDashboard } from "../../hooks/useDashboard";
import { get } from "lodash";
import { Link } from "react-router-dom";
import { useEffect, useState } from "react";

interface OtherModDashboardsDictionary {
  [key: string]: AvailableDashboard[];
}

const ModSection = ({ mod, dashboards }) => {
  return (
    <>
      <h3 className="text-xl font-medium">{mod.title || mod.short_name}</h3>
      <ul className="list-none list-inside">
        {dashboards.map((dashboard) => (
          <li key={dashboard.full_name} className="pl-4">
            <Link className="link-highlight" to={`/${dashboard.full_name}`}>
              {dashboard.title || dashboard.short_name}
            </Link>
          </li>
        ))}
      </ul>
    </>
  );
};

const CurrentModSection = ({ dashboards, metadata }) => {
  const mod = get(metadata, "mod", {});
  return <ModSection mod={mod} dashboards={dashboards} />;
};

const OtherModSection = ({ mod_full_name, dashboards, metadata }) => {
  const mod = get(metadata, `installed_mods["${mod_full_name}"]`, {});
  return <ModSection mod={mod} dashboards={dashboards} />;
};

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
  const [dashboardsForOtherMods, setDashboardsForOtherMods] =
    useState<OtherModDashboardsDictionary>({});

  useEffect(() => {
    if (!metadataLoaded || !availableDashboardsLoaded) {
      setDashboardsForCurrentMod([]);
      setDashboardsForOtherMods({});
      return;
    }

    setDashboardsForCurrentMod(
      dashboards.filter(
        (dashboard) => dashboard.mod_full_name === metadata.mod.full_name
      )
    );

    const newOtherMods = {};
    for (const [mod_full_name, mod] of Object.entries(
      metadata.installed_mods || {}
    )) {
      newOtherMods[mod_full_name] = dashboards
        .filter((dashboard) => dashboard.mod_full_name === mod_full_name)
        .map((dashboard) => ({ ...dashboard, mod }));
    }
    setDashboardsForOtherMods(newOtherMods);
  }, [metadataLoaded, availableDashboardsLoaded, metadata, dashboards]);

  if (selectedDashboard) {
    return null;
  }

  return (
    <div className="w-full p-4">
      <h2 className="text-2xl font-medium">Available Dashboards</h2>
      {!availableDashboardsLoaded && !metadataLoaded && (
        <div className="mt-2 text-black-scale-3">
          <LoadingIndicator /> <span className="italic">Loading...</span>
        </div>
      )}
      {availableDashboardsLoaded && metadataLoaded && (
        <div className="mt-4 space-y-2">
          <CurrentModSection
            dashboards={dashboardsForCurrentMod}
            metadata={metadata}
          />
          {Object.entries(dashboardsForOtherMods).map(
            ([mod_full_name, dashboards]) => (
              <OtherModSection
                key={mod_full_name}
                mod_full_name={mod_full_name}
                dashboards={dashboards}
                metadata={metadata}
              />
            )
          )}
        </div>
      )}
    </div>
  );
};

export default DashboardList;
