import LoadingIndicator from "../dashboards/LoadingIndicator";
import SearchInput from "../SearchInput";
import SlackCommunityCallToAction from "../CallToAction/SlackCommunityCallToAction";
import useDebouncedEffect from "../../hooks/useDebouncedEffect";
import useQueryParam, {
  urlQueryParamHistoryMode,
} from "../../hooks/useQueryParam";
import {
  AvailableDashboard,
  ModDashboardMetadata,
  useDashboard,
} from "../../hooks/useDashboard";
import { get, sortBy } from "lodash";
import { Link, useParams } from "react-router-dom";
import { useEffect, useState } from "react";

interface OtherModDashboardsDictionary {
  [key: string]: AvailableDashboard[];
}

const ModSection = ({ mod, dashboards }) => {
  return (
    <div className="space-y-2">
      <h3 className="truncate">{mod.title || mod.short_name}</h3>
      <ul className="list-none list-inside">
        {dashboards.map((dashboard) => (
          <li key={dashboard.full_name} className="mt-1 truncate">
            <Link className="link-highlight" to={`/${dashboard.full_name}`}>
              {dashboard.title || dashboard.short_name}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
};

const CurrentModSection = ({ dashboards, metadata }) => {
  if (dashboards.length === 0) {
    return null;
  }
  const mod = get(metadata, "mod", {});
  return <ModSection mod={mod} dashboards={dashboards} />;
};

const OtherModSection = ({ mod_full_name, dashboards, metadata }) => {
  if (dashboards.length === 0) {
    return null;
  }

  const mod = get(metadata, `installed_mods["${mod_full_name}"]`, {});
  return <ModSection mod={mod} dashboards={dashboards} />;
};

const searchAgainstDashboard = (
  dashboard: AvailableDashboard,
  mod: ModDashboardMetadata,
  searchParts: string[]
): boolean => {
  const joined = `${mod.title || ""}.${mod.short_name || ""}.${
    dashboard.title || ""
  }.${dashboard.short_name || ""}`.toLowerCase();
  return searchParts.every((searchPart) => joined.indexOf(searchPart) >= 0);
};

const sortDashboards = (dashboards: AvailableDashboard[] = []) => {
  return sortBy(dashboards, [(d) => (d.title || d.short_name).toLowerCase()]);
};

const DashboardList = () => {
  const { availableDashboardsLoaded, metadataLoaded, metadata, dashboards } =
    useDashboard();
  const [dashboardsForCurrentMod, setDashboardsForCurrentMod] = useState<
    AvailableDashboard[]
  >([]);
  const [dashboardsForOtherMods, setDashboardsForOtherMods] =
    useState<OtherModDashboardsDictionary>({});
  const [filteredDashboardsForCurrentMod, setFilteredDashboardsForCurrentMod] =
    useState(dashboardsForCurrentMod);
  const [filteredDashboardsForOtherMods, setFilteredDashboardsForOtherMods] =
    useState(dashboardsForOtherMods);
  const [searchQuery, setSearchQuery] = useQueryParam(
    "search",
    "",
    urlQueryParamHistoryMode.REPLACE
  );
  const [search, setSearch] = useState(searchQuery);
  const { dashboardName } = useParams();

  useDebouncedEffect(
    () => {
      setSearchQuery(search);
    },
    250,
    [search]
  );

  useEffect(() => {
    if (!metadataLoaded || !availableDashboardsLoaded) {
      setDashboardsForCurrentMod([]);
      setDashboardsForOtherMods({});
      return;
    }

    setDashboardsForCurrentMod(
      sortDashboards(
        dashboards.filter(
          (dashboard) => dashboard.mod_full_name === metadata.mod.full_name
        )
      )
    );

    const newOtherMods = {};
    for (const [mod_full_name, mod] of Object.entries(
      metadata.installed_mods || {}
    )) {
      newOtherMods[mod_full_name] = sortDashboards(
        dashboards
          .filter((dashboard) => dashboard.mod_full_name === mod_full_name)
          .map((dashboard) => ({ ...dashboard, mod }))
      );
    }
    setDashboardsForOtherMods(newOtherMods);
  }, [metadataLoaded, availableDashboardsLoaded, metadata, dashboards]);

  useEffect(() => {
    if (!search) {
      setFilteredDashboardsForCurrentMod(dashboardsForCurrentMod);
      setFilteredDashboardsForOtherMods(dashboardsForOtherMods);
      return;
    }

    const searchParts = search.trim().toLowerCase().split(" ");
    const filteredCurrent: AvailableDashboard[] = [];
    const filteredOther: OtherModDashboardsDictionary = {};

    dashboardsForCurrentMod.forEach((dashboard) => {
      const mod: ModDashboardMetadata = get(
        metadata,
        "mod",
        {} as ModDashboardMetadata
      );
      const include = searchAgainstDashboard(dashboard, mod, searchParts);
      if (include) {
        filteredCurrent.push(dashboard);
      }
    });

    Object.entries(dashboardsForOtherMods).forEach(
      ([mod_full_name, dashboards]) => {
        const mod: ModDashboardMetadata = get(
          metadata,
          `installed_mods["${mod_full_name}"]`,
          {} as ModDashboardMetadata
        );
        dashboards.forEach((dashboard) => {
          const include = searchAgainstDashboard(dashboard, mod, searchParts);
          if (include) {
            filteredOther[mod_full_name] = filteredOther[mod_full_name] || [];
            filteredOther[mod_full_name].push(dashboard);
          }
        });
      }
    );

    setFilteredDashboardsForCurrentMod(filteredCurrent);
    setFilteredDashboardsForOtherMods(filteredOther);
  }, [dashboardsForCurrentMod, dashboardsForOtherMods, search]);

  // Clear search after we choose a report
  useEffect(() => {
    if (dashboardName) {
      setSearch("");
    }
  }, [dashboardName]);

  if (dashboardName) {
    return null;
  }

  return (
    <div className="w-full grid grid-cols-6 p-4 gap-x-4">
      <div className="col-span-6 lg:col-span-2 space-y-4">
        <div className="mt-2">
          <SearchInput
            //@ts-ignore
            disabled={!metadataLoaded || !availableDashboardsLoaded}
            placeholder="Search dashboards..."
            value={search}
            setValue={setSearch}
          />
        </div>
        {!availableDashboardsLoaded && !metadataLoaded && (
          <div className="mt-2 text-black-scale-3">
            <LoadingIndicator /> <span className="italic">Loading...</span>
          </div>
        )}
        {filteredDashboardsForCurrentMod.length === 0 &&
          Object.keys(filteredDashboardsForOtherMods).length === 0 && (
            <div className="mt-2">
              {search ? (
                <>
                  <span>No search results.</span>{" "}
                  <span
                    className="link-highlight"
                    onClick={() => setSearch("")}
                  >
                    Clear
                  </span>
                  .
                </>
              ) : (
                <span>No dashboards defined.</span>
              )}
            </div>
          )}
        <div className="mt-4 space-y-4">
          <CurrentModSection
            dashboards={filteredDashboardsForCurrentMod}
            metadata={metadata}
          />
          {sortBy(Object.entries(filteredDashboardsForOtherMods), [
            ([mod_full_name, dashboards]) => {
              const mod = get(
                metadata,
                `installed_mods["${mod_full_name}"]`,
                {}
              );
              return mod.title || mod.short_name;
            },
          ]).map(([mod_full_name, dashboards]) => (
            <OtherModSection
              key={mod_full_name}
              mod_full_name={mod_full_name}
              dashboards={dashboards}
              metadata={metadata}
            />
          ))}
        </div>
      </div>
      <div className="hidden lg:block col-span-2" />
      <div className="col-span-6 lg:col-span-2 mt-4 lg:mt-2">
        <div className="space-y-4">
          <SlackCommunityCallToAction />
        </div>
      </div>
    </div>
  );
};

export default DashboardList;
