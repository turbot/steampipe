import SearchInput from "../SearchInput";
import usePrevious from "../../hooks/usePrevious";
import { useDashboard } from "../../hooks/useDashboard";
import { useEffect } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { useBreakpoint } from "../../hooks/useBreakpoint";

interface DashboardSearchStates {
  dashboardName: string | null;
  dashboardSearch: string | null;
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

  // Keep track of the previous selected dashboard and inputs
  const previousDashboardSearchStates: DashboardSearchStates | undefined =
    usePrevious({
      dashboardName,
      dashboardSearch,
    });

  useEffect(() => {
    const previous = previousDashboardSearchStates
      ? // @ts-ignore
        previousDashboardSearchStates.dashboardName
      : null;
    if (
      (!previous && dashboardName) ||
      (previous && previous !== dashboardName)
    ) {
      setDashboardSearch("");
    }
  }, [dashboardName, previousDashboardSearchStates]);

  useEffect(() => {
    // console.log(searchParams.get("search"), dashboardSearch);
    if (!searchParams.get("search") && !!dashboardSearch) {
      setDashboardSearch("");
    }
  }, [dashboardSearch, searchParams]);

  return (
    <div className="w-32 sm:w-56 md:w-72 lg:w-96">
      <SearchInput
        //@ts-ignore
        disabled={!metadataLoaded || !availableDashboardsLoaded}
        placeholder={minBreakpoint("md") ? "Search dashboards..." : "Search..."}
        value={dashboardSearch}
        setValue={setDashboardSearch}
      />
    </div>
  );
};

export default DashboardSearch;
