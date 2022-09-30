import {
  CheckDisplayGroup,
  CheckDisplayGroupType,
} from "../components/dashboards/check/common";
import { useSearchParams } from "react-router-dom";
import { useMemo } from "react";

const useCheckGroupingConfig = () => {
  const [searchParams] = useSearchParams();
  const groupingsConfig = useMemo(() => {
    const rawGrouping = searchParams.get("grouping");
    if (rawGrouping) {
      const groupings: CheckDisplayGroup[] = [];
      const groupingParts = rawGrouping.split(",");
      for (const groupingPart of groupingParts) {
        const typeValueParts = groupingPart.split("|");
        if (typeValueParts.length > 1) {
          groupings.push({
            type: typeValueParts[0] as CheckDisplayGroupType,
            value: typeValueParts[1],
          });
        } else {
          groupings.push({
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

  return groupingsConfig;
};

export default useCheckGroupingConfig;
