import { DashboardDataMode } from "../../types/dashboard";
import { useEffect, useState } from "react";

const buildDashboardDataModeFromSearchParams = (
  searchParams: URLSearchParams
): DashboardDataMode => {
  if (searchParams.has("mode")) {
    return searchParams.get("mode") as DashboardDataMode;
  }
  return "live";
};

const useDashboardDataMode = (searchParams: URLSearchParams) => {
  const [dataMode, setDataMode] = useState<DashboardDataMode>(
    buildDashboardDataModeFromSearchParams(searchParams)
  );

  useEffect(() => {
    setDataMode(buildDashboardDataModeFromSearchParams(searchParams));
  }, [searchParams, setDataMode]);

  return { dataMode };
};

export default useDashboardDataMode;
