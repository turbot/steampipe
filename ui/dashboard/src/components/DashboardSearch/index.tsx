import SearchInput from "../SearchInput";
import usePrevious from "../../hooks/usePrevious";
import { useDashboard } from "../../hooks/useDashboard";
import { useEffect } from "react";
import { useParams, useSearchParams } from "react-router-dom";

interface DashboardSearchStates {
  dashboardName: string | null;
}

const DashboardSearch = () => {
  const {
    availableDashboardsLoaded,
    dashboardSearch,
    metadataLoaded,
    setDashboardSearch,
  } = useDashboard();
  const { dashboardName } = useParams();
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

  return (
    <SearchInput
      //@ts-ignore
      disabled={!metadataLoaded || !availableDashboardsLoaded}
      placeholder="Search dashboards..."
      value={dashboardSearch}
      setValue={setDashboardSearch}
    />
  );
};

export default DashboardSearch;
