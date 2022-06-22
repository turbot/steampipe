import { DashboardInputs } from "../../types/dashboard";
import { useEffect, useState } from "react";

const buildDashboardInputsFromSearchParams = (
  searchParams: URLSearchParams
) => {
  const selectedDashboardInputs = {};
  // @ts-ignore
  for (const entry of searchParams.entries()) {
    if (!entry[0].startsWith("input")) {
      continue;
    }
    selectedDashboardInputs[entry[0]] = entry[1];
  }
  return selectedDashboardInputs;
};

const useDashboardInputs = (searchParams: URLSearchParams) => {
  const [inputs, setInputs] = useState<DashboardInputs>(
    buildDashboardInputsFromSearchParams(searchParams)
  );

  useEffect(() => {
    setInputs(buildDashboardInputsFromSearchParams(searchParams));
  }, [searchParams, setInputs]);

  return { inputs };
};

export default useDashboardInputs;
