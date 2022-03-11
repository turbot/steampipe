import SearchInput from "../SearchInput";
import usePrevious from "../../hooks/usePrevious";
import { DashboardActions, useDashboard } from "../../hooks/useDashboard";
import { useBreakpoint } from "../../hooks/useBreakpoint";
import { useCallback, useEffect } from "react";
import { useParams } from "react-router-dom";

interface DashboardSearchStates {
  dashboardName: string | null;
  searchParamsSearch: string | null;
}

const DashboardSearch = () => {
  const { availableDashboardsLoaded, dispatch, search, metadata } =
    useDashboard();
  const { dashboardName } = useParams();
  const { minBreakpoint } = useBreakpoint();

  const updateSearchValue = useCallback(
    (value) =>
      dispatch({ type: DashboardActions.SET_DASHBOARD_SEARCH_VALUE, value }),
    [dispatch]
  );

  // // Keep track of the previous selected dashboard and inputs
  // const previousDashboardSearchStates: DashboardSearchStates | undefined =
  //   usePrevious({
  //     dashboardName,
  //     searchParamsSearch: searchParams.get("search"),
  //   });
  //
  // useEffect(() => {
  //   const previousSearchParamsSearch = previousDashboardSearchStates
  //     ? // @ts-ignore
  //       previousDashboardSearchStates.searchParamsSearch
  //     : null;
  //
  //   if (!!previousSearchParamsSearch && !searchParams.get("search")) {
  //     updateSearchValue("");
  //   }
  // }, [previousDashboardSearchStates, searchParams, updateSearchValue]);
  //
  // useEffect(() => {
  //   const previousDashboard = previousDashboardSearchStates
  //     ? // @ts-ignore
  //       previousDashboardSearchStates.dashboardName
  //     : null;
  //   if (
  //     (!previousDashboard && dashboardName) ||
  //     (previousDashboard && previousDashboard !== dashboardName)
  //   ) {
  //     updateSearchValue("");
  //   }
  // }, [dashboardName, previousDashboardSearchStates, updateSearchValue]);
  //
  // useEffect(() => {
  //   if (!searchParams.get("search") && !!search.value) {
  //     updateSearchValue("");
  //   }
  // }, [search.value, searchParams, updateSearchValue]);
  //
  // /*eslint-disable */
  // useEffect(() => {
  //   if (search.value) {
  //     searchParams.set("search", search.value);
  //   } else {
  //     searchParams.delete("search");
  //   }
  //   setSearchParams(searchParams, { replace: true });
  // }, [search.value, searchParams]);
  // /*eslint-enable */

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
