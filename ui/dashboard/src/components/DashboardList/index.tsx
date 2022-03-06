import LoadingIndicator from "../dashboards/LoadingIndicator";
import SlackCommunityCallToAction from "../CallToAction/SlackCommunityCallToAction";
import {
  AvailableDashboard,
  ModDashboardMetadata,
  useDashboard,
} from "../../hooks/useDashboard";
import { ColorGenerator } from "../../utils/color";
import { get, groupBy as lodashGroupBy, sortBy } from "lodash";
import { Link, useParams, useSearchParams } from "react-router-dom";
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
  searchParams: URLSearchParams;
  setDashboardSearch: (search: string) => void;
}

interface SectionProps {
  title: string;
  dashboards: AvailableDashboardWithMod[];
  searchParams: URLSearchParams;
  setDashboardSearch: (search: string) => void;
}

// /*!
//  * Get the contrasting color for any hex color
//  * (c) 2019 Chris Ferdinandi, MIT License, https://gomakethings.com
//  * Derived from work by Brian Suda, https://24ways.org/2010/calculating-color-contrast/
//  * @param  {String} A hexcolor value
//  * @return {String} The contrasting color (black or white)
//  */
// // https://gomakethings.com/dynamically-changing-the-text-color-based-on-background-color-contrast-with-vanilla-js/
// const getContrastColour = (hexcolor) => {
//   // If a leading # is provided, remove it
//   if (hexcolor.slice(0, 1) === "#") {
//     hexcolor = hexcolor.slice(1);
//   }
//
//   // If a three-character hexcode, make six-character
//   if (hexcolor.length === 3) {
//     hexcolor = hexcolor
//       .split("")
//       .map(function (hex) {
//         return hex + hex;
//       })
//       .join("");
//   }
//
//   // Convert to RGB value
//   const r = parseInt(hexcolor.substr(0, 2), 16);
//   const g = parseInt(hexcolor.substr(2, 2), 16);
//   const b = parseInt(hexcolor.substr(4, 2), 16);
//
//   // Get YIQ ratio
//   const yiq = (r * 299 + g * 587 + b * 114) / 1000;
//
//   // Check contrast
//   return yiq >= 128 ? "black" : "white";
// };

const stringColorMap = {};
const colorGenerator = new ColorGenerator(16, 0);

// https://stackoverflow.com/questions/3426404/create-a-hexadecimal-colour-based-on-a-string-with-javascript
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
  searchParams,
  setDashboardSearch,
}: DashboardTagProps) => {
  const search = searchParams.get("search");
  const searchWithTag = useMemo(() => {
    const existingSearch = (search || "").trim();
    return existingSearch
      ? existingSearch.indexOf(tagValue) < 0
        ? `${existingSearch} ${tagValue}`
        : existingSearch
      : tagValue;
  }, [tagValue, search]);

  return (
    <span
      className="cursor-pointer rounded-md text-xs"
      onClick={() => setDashboardSearch(searchWithTag)}
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
  searchParams,
  setDashboardSearch,
}: SectionProps) => {
  return (
    <div className="space-y-2">
      <h3 className="truncate">{title}</h3>
      {dashboards.map((dashboard) => (
        <div key={dashboard.full_name} className="flex space-x-2 items-center">
          <div className="md:col-span-6 truncate">
            <Link className="link-highlight" to={`/${dashboard.full_name}`}>
              {dashboard.title || dashboard.short_name}
            </Link>
          </div>
          <div className="hidden md:block col-span-6 space-x-2">
            {Object.entries(dashboard.tags || {}).map(([key, value]) => (
              <DashboardTag
                key={key}
                tagKey={key}
                tagValue={value}
                searchParams={searchParams}
                setDashboardSearch={setDashboardSearch}
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

const useGroupedDashboards = (dashboards, group_by, tag, metadata) => {
  const [sections, setSections] = useState<DashboardListSection[]>([]);

  useEffect(() => {
    let groupedDashboards: GroupedDashboards;
    if (group_by === "tag") {
      groupedDashboards = lodashGroupBy(dashboards, (dashboard) => {
        return get(dashboard, `tags["${tag}"]`, "Other");
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
  }, [dashboards, group_by, tag, metadata]);

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
  const [searchParams] = useSearchParams();
  const {
    availableDashboardsLoaded,
    metadataLoaded,
    metadata,
    dashboards,
    dashboardSearch: search,
    setDashboardSearch,
    setDashboardTagKeys,
  } = useDashboard();
  const [unfilteredDashboards, setUnfilteredDashboards] = useState<
    AvailableDashboardWithMod[]
  >([]);
  const [filteredDashboards, setFilteredDashboards] = useState<
    AvailableDashboardWithMod[]
  >([]);

  // Initialise dashboards with their mod + update when the list of dashboards is updated
  useEffect(() => {
    if (!availableDashboardsLoaded) {
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
    setDashboardTagKeys(newDashboardTagKeys);
  }, [availableDashboardsLoaded, dashboards, metadata]);

  // Filter dashboards according to the search
  useEffect(() => {
    if (!availableDashboardsLoaded || !metadataLoaded) {
      return;
    }
    if (!search) {
      setFilteredDashboards(unfilteredDashboards);
      return;
    }

    const searchParts = search.trim().toLowerCase().split(" ");
    const filtered: AvailableDashboard[] = [];

    unfilteredDashboards.forEach((dashboard) => {
      const include = searchAgainstDashboard(dashboard, searchParts);
      if (include) {
        filtered.push(dashboard);
      }
    });

    setFilteredDashboards(sortDashboards(filtered));
  }, [
    availableDashboardsLoaded,
    metadataLoaded,
    unfilteredDashboards,
    metadata,
    search,
  ]);

  const url_group_by = searchParams.get("group_by") || "tag";
  const url_tag = searchParams.get("tag") || "service";

  const sections = useGroupedDashboards(
    filteredDashboards,
    url_group_by,
    url_tag,
    metadata
  );

  return (
    <div className="w-full grid grid-cols-6 p-4 gap-x-4">
      <div className="col-span-6 lg:col-span-4 space-y-4">
        <div className="grid grid-cols-6">
          {(!availableDashboardsLoaded || !metadataLoaded) && (
            <div className="col-span-6 mt-4 ml-1 text-black-scale-4 flex">
              <LoadingIndicator className="w-4 h-4" />{" "}
              <span className="italic -ml-1">Loading...</span>
            </div>
          )}
          <div className="col-span-6">
            {availableDashboardsLoaded &&
              metadataLoaded &&
              filteredDashboards.length === 0 && (
                <div className="col-span-6 mt-4">
                  {search ? (
                    <>
                      <span>No search results.</span>{" "}
                      <span
                        className="link-highlight"
                        onClick={() => setDashboardSearch("")}
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
            <div className="mt-2 space-y-4">
              {sections.map((section) => (
                <Section
                  key={section.title}
                  title={section.title}
                  dashboards={section.dashboards}
                  searchParams={searchParams}
                  setDashboardSearch={setDashboardSearch}
                />
              ))}
            </div>
          </div>
        </div>
      </div>
      <div className="col-span-6 lg:col-span-2 mt-4 lg:mt-2">
        <div className="space-y-4">
          <SlackCommunityCallToAction />
        </div>
      </div>
    </div>
  );
};

const DashboardListWrapper = () => {
  const { dashboardName } = useParams();
  const { dashboardSearch } = useDashboard();

  // If we have a dashboard selected and no search, we don't want to show the list
  if (dashboardName && !dashboardSearch) {
    return null;
  }

  return <DashboardList />;
};

export default DashboardListWrapper;
