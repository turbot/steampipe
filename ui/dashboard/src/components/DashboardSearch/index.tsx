import SearchInput from "../SearchInput";
import { DashboardActions, useDashboard } from "../../hooks/useDashboard";
import { useCallback } from "react";

const DashboardSearch = () => {
  const {
    availableDashboardsLoaded,
    breakpointContext: { minBreakpoint },
    dispatch,
    search,
    metadata,
  } = useDashboard();

  const updateSearchValue = useCallback(
    (value) =>
      dispatch({ type: DashboardActions.SET_DASHBOARD_SEARCH_VALUE, value }),
    [dispatch]
  );

  return (
    <div className="w-32 sm:w-56 md:w-72 lg:w-96">
      <SearchInput
        //@ts-ignore
        disabled={!metadata || !availableDashboardsLoaded}
        placeholder={minBreakpoint("sm") ? "Search dashboards..." : "Search..."}
        value={search.value}
        setValue={updateSearchValue}
      />
    </div>
  );
};

export default DashboardSearch;
