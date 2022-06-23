import SearchInput from "../SearchInput";
import { useDashboardNew } from "../../hooks/refactor/useDashboard";
import { useSearchParams } from "react-router-dom";

const DashboardSearch = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const {
    availableDashboardsLoaded,
    breakpointContext: { minBreakpoint },
    metadata,
    search,
  } = useDashboardNew();

  const updateSearchValue = (value) => {
    if (value) {
      searchParams.set("search", value);
    } else {
      searchParams.delete("search");
    }
    setSearchParams(searchParams, { replace: true });
  };

  return (
    <div className="w-full sm:w-56 md:w-72 lg:w-96">
      <SearchInput
        disabled={!metadata || !availableDashboardsLoaded}
        placeholder={minBreakpoint("sm") ? "Search dashboards..." : "Search..."}
        value={search.value}
        setValue={updateSearchValue}
      />
    </div>
  );
};

export default DashboardSearch;
