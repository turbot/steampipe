import LoadingIndicator from "../dashboards/LoadingIndicator";
import SearchInput from "../SearchInput";
import SlackCommunityCallToAction from "../CallToAction/SlackCommunityCallToAction";
import useDebouncedEffect from "../../hooks/useDebouncedEffect";
import {
  AvailableDashboard,
  ModDashboardMetadata,
  useDashboard,
} from "../../hooks/useDashboard";
import { get, groupBy, sortBy } from "lodash";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { useEffect, useState } from "react";

interface DashboardListSection {
  title: string;
  dashboards: AvailableDashboardWithMod[];
}

type AvailableDashboardWithMod = AvailableDashboard & {
  mod?: ModDashboardMetadata;
};

const Section = ({ title, dashboards }) => {
  return (
    <div className="space-y-2">
      <h3 className="truncate">{title}</h3>
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

// const CurrentModSection = ({ dashboards, metadata }) => {
//   if (dashboards.length === 0) {
//     return null;
//   }
//   const mod = get(metadata, "mod", {});
//   return (
//     <Section title={mod.title || mod.short_name} dashboards={dashboards} />
//   );
// };
//
// const OtherModSection = ({ mod_full_name, dashboards, metadata }) => {
//   if (dashboards.length === 0) {
//     return null;
//   }
//
//   const mod = get(metadata, `installed_mods["${mod_full_name}"]`, {});
//   return (
//     <Section title={mod.title || mod.short_name} dashboards={dashboards} />
//   );
// };

interface GroupedDashboards {
  [key: string]: AvailableDashboardWithMod[];
}

const useGroupedDashboards = (dashboards, grouping, metadata) => {
  const [sections, setSections] = useState<DashboardListSection[]>([]);

  useEffect(() => {
    let groupedDashboards: GroupedDashboards;
    if (grouping === "mod") {
      groupedDashboards = groupBy(dashboards, (dashboard) => {
        return get(
          dashboard,
          `mod.title`,
          get(dashboard, "mod.short_name", "Other")
        );
      });
    } else {
      groupedDashboards = groupBy(dashboards, (dashboard) => {
        return get(dashboard, `tags["${grouping}"]`, "Other");
      });
    }
    setSections(
      Object.entries(groupedDashboards).map(([k, v]) => ({
        title: k,
        dashboards: v,
      }))
    );
  }, [dashboards, grouping, metadata]);

  return sections;
};

const searchAgainstDashboard = (
  dashboard: AvailableDashboardWithMod,
  searchParts: string[]
): boolean => {
  const joined = `${dashboard.mod?.title || ""}.${
    dashboard.mod?.short_name || ""
  }.${dashboard.title || ""}.${dashboard.short_name || ""}`.toLowerCase();
  return searchParts.every((searchPart) => joined.indexOf(searchPart) >= 0);
};

const sortDashboards = (dashboards: AvailableDashboard[] = []) => {
  return sortBy(dashboards, [(d) => (d.title || d.short_name).toLowerCase()]);
};

const DashboardList = () => {
  const [searchParams, setSearchParams] = useSearchParams({
    grouping: "mod",
    search: "",
  });
  const [search, setSearch] = useState(searchParams.get("search"));
  const { availableDashboardsLoaded, metadataLoaded, metadata, dashboards } =
    useDashboard();
  const [unfilteredDashboards, setUnfilteredDashboards] = useState<
    AvailableDashboardWithMod[]
  >([]);
  const [filteredDashboards, setFilteredDashboards] = useState<
    AvailableDashboardWithMod[]
  >([]);

  const { dashboardName } = useParams();

  useDebouncedEffect(
    () => {
      if (search) {
        searchParams.set("search", search);
      } else {
        searchParams.delete("search");
      }
      setSearchParams(searchParams);
    },
    250,
    [search]
  );

  // Initialise dashboards with their mod + update when the list of dashboards is updated
  useEffect(() => {
    if (!availableDashboardsLoaded) {
      setUnfilteredDashboards([]);
      return;
    }

    const dashboardsWithMod: AvailableDashboardWithMod[] = [];

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
    }
    setUnfilteredDashboards(dashboardsWithMod);
  }, [availableDashboardsLoaded, dashboards]);

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

  // Clear search after we choose a report
  useEffect(() => {
    if (dashboardName) {
      setSearch("");
    }
  }, [dashboardName]);

  const sections = useGroupedDashboards(
    filteredDashboards,
    searchParams.get("grouping"),
    metadata
  );

  // useEffect(() => {
  //   if (!metadataLoaded || !availableDashboardsLoaded) {
  //     setDashboardsForCurrentMod([]);
  //     setDashboardsForOtherMods({});
  //     return;
  //   }
  //
  //   setDashboardsForCurrentMod(
  //     sortDashboards(
  //       dashboards.filter(
  //         (dashboard) => dashboard.mod_full_name === metadata.mod.full_name
  //       )
  //     )
  //   );
  //
  //   const newOtherMods = {};
  //   for (const [mod_full_name, mod] of Object.entries(
  //     metadata.installed_mods || {}
  //   )) {
  //     newOtherMods[mod_full_name] = sortDashboards(
  //       dashboards
  //         .filter((dashboard) => dashboard.mod_full_name === mod_full_name)
  //         .map((dashboard) => ({ ...dashboard, mod }))
  //     );
  //   }
  //   setDashboardsForOtherMods(newOtherMods);
  // }, [metadataLoaded, availableDashboardsLoaded, metadata, dashboards]);
  //
  // useEffect(() => {
  //   if (!search) {
  //     setFilteredDashboardsForCurrentMod(dashboardsForCurrentMod);
  //     setFilteredDashboardsForOtherMods(dashboardsForOtherMods);
  //     return;
  //   }
  //
  //   const searchParts = search.trim().toLowerCase().split(" ");
  //   const filteredCurrent: AvailableDashboard[] = [];
  //   const filteredOther: OtherModDashboardsDictionary = {};
  //
  //   dashboardsForCurrentMod.forEach((dashboard) => {
  //     const mod: ModDashboardMetadata = get(
  //       metadata,
  //       "mod",
  //       {} as ModDashboardMetadata
  //     );
  //     const include = searchAgainstDashboard(dashboard, mod, searchParts);
  //     if (include) {
  //       filteredCurrent.push(dashboard);
  //     }
  //   });
  //
  //   Object.entries(dashboardsForOtherMods).forEach(
  //     ([mod_full_name, dashboards]) => {
  //       const mod: ModDashboardMetadata = get(
  //         metadata,
  //         `installed_mods["${mod_full_name}"]`,
  //         {} as ModDashboardMetadata
  //       );
  //       dashboards.forEach((dashboard) => {
  //         const include = searchAgainstDashboard(dashboard, mod, searchParts);
  //         if (include) {
  //           filteredOther[mod_full_name] = filteredOther[mod_full_name] || [];
  //           filteredOther[mod_full_name].push(dashboard);
  //         }
  //       });
  //     }
  //   );
  //
  //   setFilteredDashboardsForCurrentMod(filteredCurrent);
  //   setFilteredDashboardsForOtherMods(filteredOther);
  // }, [dashboardsForCurrentMod, dashboardsForOtherMods, metadata, search]);

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
        {(!availableDashboardsLoaded || !metadataLoaded) && (
          <div className="mt-2 ml-1 text-black-scale-4 flex">
            <LoadingIndicator className="w-4 h-4" />{" "}
            <span className="italic -ml-1">Loading...</span>
          </div>
        )}
        {availableDashboardsLoaded &&
          metadataLoaded &&
          filteredDashboards.length === 0 && (
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
          {sections.map((section) => (
            <Section
              key={section.title}
              title={section.title}
              dashboards={section.dashboards}
            />
          ))}

          {/*<CurrentModSection*/}
          {/*  dashboards={filteredDashboardsForCurrentMod}*/}
          {/*  metadata={metadata}*/}
          {/*/>*/}
          {/*{sortBy(Object.entries(filteredDashboardsForOtherMods), [*/}
          {/*  ([mod_full_name, dashboards]) => {*/}
          {/*    const mod = get(*/}
          {/*      metadata,*/}
          {/*      `installed_mods["${mod_full_name}"]`,*/}
          {/*      {}*/}
          {/*    );*/}
          {/*    return mod.title || mod.short_name;*/}
          {/*  },*/}
          {/*]).map(([mod_full_name, dashboards]) => (*/}
          {/*  <OtherModSection*/}
          {/*    key={mod_full_name}*/}
          {/*    mod_full_name={mod_full_name}*/}
          {/*    dashboards={dashboards}*/}
          {/*    metadata={metadata}*/}
          {/*  />*/}
          {/*))}*/}
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
