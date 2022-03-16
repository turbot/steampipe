import {
  AvailableDashboard,
  DashboardAction,
  DashboardActions,
  ModDashboardMetadata,
  useDashboard,
} from "../../hooks/useDashboard";
import CallToActions from "../CallToActions";
import LoadingIndicator from "../dashboards/LoadingIndicator";
import { ColorGenerator } from "../../utils/color";
import { get, groupBy as lodashGroupBy, sortBy } from "lodash";
import { Link, useParams } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";

interface DashboardListSection {
  title: string;
  dashboards: AvailableDashboardWithMod[];
}

type AvailableDashboardWithMod = AvailableDashboard & {
  mod?: ModDashboardMetadata;
};

interface DashboardTagProps {
  tagKey: string;
  tagValue: string;
  dispatch: (action: DashboardAction) => void;
  searchValue: string;
}

interface SectionProps {
  title: string;
  dashboards: AvailableDashboardWithMod[];
  dispatch: (action: DashboardAction) => void;
  searchValue: string;
}

const stringColorMap = {};
const colorGenerator = new ColorGenerator(16, 0);

const stringToColour = (str) => {
  if (stringColorMap[str]) {
    return stringColorMap[str];
  }
  const color = colorGenerator.nextColor().hex;
  stringColorMap[str] = color;
  return color;
};

const DashboardTag = ({
  tagKey,
  tagValue,
  dispatch,
  searchValue,
}: DashboardTagProps) => {
  const searchWithTag = useMemo(() => {
    const existingSearch = searchValue.trim();
    return existingSearch
      ? existingSearch.indexOf(tagValue) < 0
        ? `${existingSearch} ${tagValue}`
        : existingSearch
      : tagValue;
  }, [tagValue, searchValue]);

  return (
    <span
      className="cursor-pointer rounded-md text-xs"
      onClick={() =>
        dispatch({
          type: DashboardActions.SET_DASHBOARD_SEARCH_VALUE,
          value: searchWithTag,
        })
      }
      style={{ color: stringToColour(tagValue) }}
      title={`${tagKey} = ${tagValue}`}
    >
      {tagValue}
    </span>
  );
};

const Section = ({
  title,
  dashboards,
  dispatch,
  searchValue,
}: SectionProps) => {
  return (
    <div className="space-y-2">
      <h3 className="truncate">{title}</h3>
      {dashboards.map((dashboard) => (
        <div key={dashboard.full_name} className="flex space-x-2 items-center">
          <div className="md:col-span-6 truncate">
            <Link className="link-highlight" to={dashboard.full_name}>
              {dashboard.title || dashboard.short_name}
            </Link>
          </div>
          <div className="hidden md:block col-span-6 space-x-2">
            {Object.entries(dashboard.tags || {}).map(([key, value]) => (
              <DashboardTag
                key={key}
                tagKey={key}
                tagValue={value}
                dispatch={dispatch}
                searchValue={searchValue}
              />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
};

interface GroupedDashboards {
  [key: string]: AvailableDashboardWithMod[];
}

const useGroupedDashboards = (dashboards, group_by, metadata) => {
  const [sections, setSections] = useState<DashboardListSection[]>([]);

  useEffect(() => {
    let groupedDashboards: GroupedDashboards;
    if (group_by.value === "tag") {
      groupedDashboards = lodashGroupBy(dashboards, (dashboard) => {
        return get(dashboard, `tags["${group_by.tag}"]`, "Other");
      });
    } else {
      groupedDashboards = lodashGroupBy(dashboards, (dashboard) => {
        return get(
          dashboard,
          `mod.title`,
          get(dashboard, "mod.short_name", "Other")
        );
      });
    }
    setSections(
      Object.entries(groupedDashboards)
        .map(([k, v]) => ({
          title: k,
          dashboards: v,
        }))
        .sort((x, y) => {
          if (y.title === "Other") {
            return -1;
          }
          if (x.title < y.title) {
            return -1;
          }
          if (x.title > y.title) {
            return 1;
          }
          return 0;
        })
    );
  }, [dashboards, group_by, metadata]);

  return sections;
};

const searchAgainstDashboard = (
  dashboard: AvailableDashboardWithMod,
  searchParts: string[]
): boolean => {
  const joined = `${dashboard.mod?.title || dashboard.mod?.short_name || ""} ${
    dashboard.title || dashboard.short_name || ""
  } ${Object.entries(dashboard.tags || {})
    .map(([tagKey, tagValue]) => `${tagKey}=${tagValue}`)
    .join(" ")}`.toLowerCase();
  return searchParts.every((searchPart) => joined.indexOf(searchPart) >= 0);
};

const sortDashboards = (dashboards: AvailableDashboard[] = []) => {
  return sortBy(dashboards, [(d) => (d.title || d.short_name).toLowerCase()]);
};

const DashboardList = () => {
  const {
    availableDashboardsLoaded,
    metadata,
    dashboards,
    dispatch,
    search: { value: searchValue, groupBy: searchGroupBy },
  } = useDashboard();
  const [unfilteredDashboards, setUnfilteredDashboards] = useState<
    AvailableDashboardWithMod[]
  >([]);
  const [filteredDashboards, setFilteredDashboards] = useState<
    AvailableDashboardWithMod[]
  >([]);

  // Initialise dashboards with their mod + update when the list of dashboards is updated
  useEffect(() => {
    if (!metadata || !availableDashboardsLoaded) {
      setUnfilteredDashboards([]);
      return;
    }

    const dashboardsWithMod: AvailableDashboardWithMod[] = [];
    const newDashboardTagKeys: string[] = [];
    for (const dashboard of dashboards) {
      const dashboardMod = dashboard.mod_full_name;
      let mod: ModDashboardMetadata;
      if (dashboardMod === metadata.mod.full_name) {
        mod = get(metadata, "mod", {} as ModDashboardMetadata);
      } else {
        mod = get(
          metadata,
          `installed_mods["${dashboardMod}"]`,
          {} as ModDashboardMetadata
        );
      }
      let dashboardWithMod: AvailableDashboardWithMod;
      dashboardWithMod = { ...dashboard };
      dashboardWithMod.mod = mod;
      dashboardsWithMod.push(dashboardWithMod);

      Object.entries(dashboard.tags || {}).forEach(([tagKey]) => {
        if (!newDashboardTagKeys.includes(tagKey)) {
          newDashboardTagKeys.push(tagKey);
        }
      });
    }
    setUnfilteredDashboards(dashboardsWithMod);
    dispatch({
      type: DashboardActions.SET_DASHBOARD_TAG_KEYS,
      keys: newDashboardTagKeys,
    });
  }, [availableDashboardsLoaded, dashboards, dispatch, metadata]);

  // Filter dashboards according to the search
  useEffect(() => {
    if (!availableDashboardsLoaded || !metadata) {
      return;
    }
    if (!searchValue) {
      setFilteredDashboards(unfilteredDashboards);
      return;
    }

    const searchParts = searchValue.trim().toLowerCase().split(" ");
    const filtered: AvailableDashboard[] = [];

    unfilteredDashboards.forEach((dashboard) => {
      const include = searchAgainstDashboard(dashboard, searchParts);
      if (include) {
        filtered.push(dashboard);
      }
    });

    setFilteredDashboards(sortDashboards(filtered));
  }, [availableDashboardsLoaded, unfilteredDashboards, metadata, searchValue]);

  const sections = useGroupedDashboards(
    filteredDashboards,
    searchGroupBy,
    metadata
  );

  return (
    <div className="w-full grid grid-cols-12 p-4 gap-x-4">
      <div className="col-span-12 lg:col-span-9 space-y-4">
        <div className="grid grid-cols-6">
          {(!availableDashboardsLoaded || !metadata) && (
            <div className="col-span-6 mt-2 ml-1 text-black-scale-4 flex">
              <LoadingIndicator className="w-4 h-4" />{" "}
              <span className="italic -ml-1">Loading...</span>
            </div>
          )}
          <div className="col-span-6">
            {availableDashboardsLoaded &&
              metadata &&
              filteredDashboards.length === 0 && (
                <div className="col-span-6 mt-2">
                  {searchValue ? (
                    <>
                      <span>No search results.</span>{" "}
                      <span
                        className="link-highlight"
                        onClick={() =>
                          dispatch({
                            type: DashboardActions.SET_DASHBOARD_SEARCH_VALUE,
                            value: "",
                          })
                        }
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
            <div className="space-y-4">
              {sections.map((section) => (
                <Section
                  key={section.title}
                  title={section.title}
                  dashboards={section.dashboards}
                  dispatch={dispatch}
                  searchValue={searchValue}
                />
              ))}
            </div>
          </div>
        </div>
      </div>
      <div className="col-span-12 lg:col-span-3 mt-4 lg:mt-2">
        <CallToActions />
      </div>
    </div>
  );
};

const DashboardListWrapper = () => {
  const { dashboardName } = useParams();
  const { search } = useDashboard();

  // If we have a dashboard selected and no search, we don't want to show the list
  if (dashboardName && !search.value) {
    return null;
  }

  return <DashboardList />;
};

export default DashboardListWrapper;
