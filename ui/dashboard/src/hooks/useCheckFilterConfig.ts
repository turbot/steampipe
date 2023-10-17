import { AndFilter, CheckFilter } from "../components/dashboards/check/common";
import { useMemo } from "react";
import { useSearchParams } from "react-router-dom";

const useCheckFilterConfig = () => {
  const [searchParams] = useSearchParams();
  return useMemo(() => {
    const rawFilters = searchParams.get("where");
    const emptyAnd: AndFilter = { and: [] };
    if (rawFilters) {
      try {
        let parsedFilters: CheckFilter;
        parsedFilters = JSON.parse(rawFilters);
        if (!parsedFilters.and) {
          parsedFilters.and = [];
        }
        return parsedFilters;
      } catch (error) {
        console.error("Error parsing where filters", error);
        return { ...emptyAnd } as CheckFilter;
      }
    } else {
      return { ...emptyAnd } as CheckFilter;
    }
  }, [searchParams]);
};

export default useCheckFilterConfig;
