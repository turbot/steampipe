import {
  CheckDisplayGroup,
  CheckDisplayGroupType,
} from "../components/dashboards/check/common";
import { useSearchParams } from "react-router-dom";
import { useMemo } from "react";

const useCheckGroupingConfig = () => {
  const [searchParams] = useSearchParams();
  return useMemo(() => {
    const rawGrouping = searchParams.get("grouping");
    if (rawGrouping) {
      const groupings: CheckDisplayGroup[] = [];
      const groupingParts = rawGrouping.split(",");
      for (const groupingPart of groupingParts) {
        const typeValueParts = groupingPart.split("|");
        if (typeValueParts.length > 1) {
          groupings.push({
            id: `${typeValueParts[0]}-${typeValueParts[1]}`,
            type: typeValueParts[0] as CheckDisplayGroupType,
            value: typeValueParts[1],
          });
        } else {
          groupings.push({
            id: typeValueParts[0],
            type: typeValueParts[0] as CheckDisplayGroupType,
          });
        }
      }
      return groupings;
    } else {
      return [
        // { type: "status" },
        // { type: "reason" },
        // { type: "resource" },
        // { type: "severity" },
        // { type: "dimension", value: "account_id" },
        // { type: "dimension", value: "region" },
        // { type: "tag", value: "service" },
        // { type: "tag", value: "cis_type" },
        // { type: "tag", value: "cis_level" },
        { type: "benchmark" },
        { type: "control" },
        { type: "result" },
      ] as CheckDisplayGroup[];
    }
  }, [searchParams]);
};

export default useCheckGroupingConfig;
