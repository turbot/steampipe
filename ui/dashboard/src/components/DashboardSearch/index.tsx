import SearchInput from "../SearchInput";
import usePrevious from "../../hooks/usePrevious";
import { useBreakpoint } from "../../hooks/useBreakpoint";
import { useDashboard } from "../../hooks/useDashboard";
import { useEffect } from "react";
import { useParams, useSearchParams } from "react-router-dom";

interface DashboardSearchStates {
  dashboardName: string | null;
  dashboardSearch: string | null;
  searchParamsSearch: string | null;
}

const DashboardSearch = () => {
  const {
    availableDashboardsLoaded,
    dashboardSearch,
    metadataLoaded,
    setDashboardSearch,
  } = useDashboard();
  const { dashboardName } = useParams();
  const { minBreakpoint } = useBreakpoint();
  const [searchParams, setSearchParams] = useSearchParams();

  // Keep track of the previous selected dashboard and inputs
  const previousDashboardSearchStates: DashboardSearchStates | undefined =
    usePrevious({
      dashboardName,
      dashboardSearch,
      searchParamsSearch: searchParams.get("search"),
    });

  useEffect(() => {
    const previousSearchParamsSearch = previousDashboardSearchStates
      ? // @ts-ignore
        previousDashboardSearchStates.searchParamsSearch
      : null;

    if (!!previousSearchParamsSearch && !searchParams.get("search")) {
      setDashboardSearch("");
    }
  }, [previousDashboardSearchStates, searchParams]);

  /*eslint-disable */
  useEffect(() => {
    if (dashboardSearch) {
      searchParams.set("search", dashboardSearch);
    } else {
      searchParams.delete("search");
    }
    setSearchParams(searchParams);
  }, [dashboardSearch, searchParams]);
  /*eslint-enable */

  useEffect(() => {
    const previousDashboard = previousDashboardSearchStates
      ? // @ts-ignore
        previousDashboardSearchStates.dashboardName
      : null;
    if (
      (!previousDashboard && dashboardName) ||
      (previousDashboard && previousDashboard !== dashboardName)
    ) {
      setDashboardSearch("");
    }
  }, [dashboardName, previousDashboardSearchStates]);

  useEffect(() => {
    if (!searchParams.get("search") && !!dashboardSearch) {
      setDashboardSearch("");
    }
  }, [dashboardSearch, searchParams]);

  return (
    <div className="w-32 sm:w-56 md:w-72 lg:w-96">
      <SearchInput
        //@ts-ignore
        disabled={!metadataLoaded || !availableDashboardsLoaded}
        placeholder={minBreakpoint("sm") ? "Search dashboards..." : "Search..."}
        value={dashboardSearch}
        setValue={setDashboardSearch}
      />
    </div>
  );
};

export default DashboardSearch;
