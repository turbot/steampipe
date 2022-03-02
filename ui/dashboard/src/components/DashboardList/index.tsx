import LoadingIndicator from "../dashboards/LoadingIndicator";
import SearchInput from "../SearchInput";
import SlackCommunityCallToAction from "../CallToAction/SlackCommunityCallToAction";
import {
  AvailableDashboard,
  ModDashboardMetadata,
  useDashboard,
} from "../../hooks/useDashboard";
import { classNames } from "../../utils/styles";
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
}

interface SectionProps {
  title: string;
  dashboards: AvailableDashboardWithMod[];
  searchParams: URLSearchParams;
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
}: DashboardTagProps) => {
  const group_by = searchParams.get("group_by");
  const tag = searchParams.get("tag");
  const search = searchParams.get("search");
  const searchUrl = useMemo(() => {
    const newSearchParams = new URLSearchParams();
    if (group_by) {
      newSearchParams.set("group_by", group_by);
    }
    if (tag) {
      newSearchParams.set("tag", tag);
    }
    const existingSearch = (search || "").trim();
    newSearchParams.set(
      "search",
      existingSearch
        ? existingSearch.indexOf(tagValue) < 0
          ? `${existingSearch} ${tagValue}`
          : existingSearch
        : tagValue
    );
    return newSearchParams.toString();
  }, [tagValue, group_by, tag, search]);

  return (
    <Link to={`/?${searchUrl}`}>
      <span
        className="rounded-md text-xs"
        style={{ color: stringToColour(tagValue) }}
        title={`${tagKey} = ${tagValue}`}
      >
        {tagValue}
      </span>
    </Link>
  );
};

const Section = ({ title, dashboards, searchParams }: SectionProps) => {
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
  const joined = `${dashboard.mod?.title || ""} ${
    dashboard.mod?.short_name || ""
  } ${dashboard.title || ""} ${dashboard.short_name || ""} ${Object.entries(
    dashboard.tags || {}
  )
    .map(([tagKey, tagValue]) => `${tagKey}=${tagValue}`)
    .join(" ")}`.toLowerCase();
  return searchParts.every((searchPart) => joined.indexOf(searchPart) >= 0);
};

const sortDashboards = (dashboards: AvailableDashboard[] = []) => {
  return sortBy(dashboards, [(d) => (d.title || d.short_name).toLowerCase()]);
};

const DashboardList = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const [search, setSearch] = useState(searchParams.get("search") || "");
  const { availableDashboardsLoaded, metadataLoaded, metadata, dashboards } =
    useDashboard();
  const [unfilteredDashboards, setUnfilteredDashboards] = useState<
    AvailableDashboardWithMod[]
  >([]);
  const [filteredDashboards, setFilteredDashboards] = useState<
    AvailableDashboardWithMod[]
  >([]);
  const [dashboardTagKeys, setDashboardTagKeys] = useState<string[]>([]);

  /*eslint-disable */
  useEffect(() => {
    if (search) {
      searchParams.set("search", search);
    } else {
      searchParams.delete("search");
    }
    setSearchParams(searchParams);
  }, [search]);
  /*eslint-enable */

  useEffect(() => {
    const newSearch = searchParams.get("search");
    setSearch(newSearch || "");
  }, [searchParams]);

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

  const { modGroupUrl, typeGroupUrl, categoryGroupUrl, serviceGroupUrl } =
    useMemo(() => {
      const url_search = searchParams.get("search");

      const modGroupSearchParams = new URLSearchParams();
      modGroupSearchParams.set("group_by", "mod");
      if (url_search) modGroupSearchParams.set("search", url_search);

      const typeGroupSearchParams = new URLSearchParams();
      typeGroupSearchParams.set("group_by", "tag");
      typeGroupSearchParams.set("tag", "type");
      if (url_search) typeGroupSearchParams.set("search", url_search);

      const categoryGroupSearchParams = new URLSearchParams();
      categoryGroupSearchParams.set("group_by", "tag");
      categoryGroupSearchParams.set("tag", "category");
      if (url_search) categoryGroupSearchParams.set("search", url_search);

      const serviceGroupSearchParams = new URLSearchParams();
      serviceGroupSearchParams.set("group_by", "tag");
      serviceGroupSearchParams.set("tag", "service");
      if (url_search) serviceGroupSearchParams.set("search", url_search);

      return {
        modGroupUrl: modGroupSearchParams.toString(),
        typeGroupUrl: typeGroupSearchParams.toString(),
        categoryGroupUrl: categoryGroupSearchParams.toString(),
        serviceGroupUrl: serviceGroupSearchParams.toString(),
      };
    }, [searchParams]);

  const url_group_by = searchParams.get("group_by") || "tag";
  const url_tag = searchParams.get("tag") || "category";

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
          <div className="col-span-6 lg:col-span-3 mt-2">
            <SearchInput
              //@ts-ignore
              disabled={!metadataLoaded || !availableDashboardsLoaded}
              placeholder="Search dashboards..."
              value={search || ""}
              setValue={setSearch}
            />
          </div>
          <div className="mt-4 col-span-6 flex flex-wrap space-x-2">
            <div>Group by:</div>
            <Link
              className={classNames(
                "block",
                url_group_by === "mod"
                  ? "text-foreground-lighter"
                  : "link-highlight"
              )}
              to={`/?${modGroupUrl}`}
            >
              Mod
            </Link>
            {dashboardTagKeys.includes("category") && (
              <Link
                className={classNames(
                  "block",
                  url_group_by === "tag" && url_tag === "category"
                    ? "text-foreground-lighter"
                    : "link-highlight"
                )}
                to={`/?${categoryGroupUrl}`}
              >
                Category
              </Link>
            )}
            {dashboardTagKeys.includes("service") && (
              <Link
                className={classNames(
                  "block",
                  url_group_by === "tag" && url_tag === "service"
                    ? "text-foreground-lighter"
                    : "link-highlight"
                )}
                to={`/?${serviceGroupUrl}`}
              >
                Service
              </Link>
            )}
            {dashboardTagKeys.includes("type") && (
              <Link
                className={classNames(
                  "block",
                  url_group_by === "tag" && url_tag === "type"
                    ? "text-foreground-lighter"
                    : "link-highlight"
                )}
                to={`/?${typeGroupUrl}`}
              >
                Type
              </Link>
            )}
          </div>
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
              {sections.map((section) => (
                <Section
                  key={section.title}
                  title={section.title}
                  dashboards={section.dashboards}
                  searchParams={searchParams}
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

  if (dashboardName) {
    return null;
  }

  return <DashboardList />;
};

export default DashboardListWrapper;
