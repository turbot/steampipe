import {
  CheckDisplayGroup,
  CheckDisplayGroupType,
} from "../components/dashboards/check/common";
import { useMemo } from "react";
import { useSearchParams } from "react-router-dom";

const groupingKeys = [
  "benchmark",
  "control",
  "dimension",
  "reason",
  "resource",
  "result",
  "severity",
  "status",
  "tag",
];

const useCheckGroupingConfig = () => {
  const [searchParams] = useSearchParams();
  return useMemo(() => {
    const rawGrouping = searchParams.get("grouping");
    if (rawGrouping) {
      const groupings: CheckDisplayGroup[] = [];
      const groupingParts = rawGrouping.split(",").filter((g) => !!g);
      for (const groupingPart of groupingParts) {
        const typeValueParts = groupingPart.split("|");
        const groupingKey = typeValueParts[0];

        // Is this a valid grouping key?
        const isValid = groupingKeys.includes(groupingKey);
        if (!isValid) {
          throw new Error(`Unsupported grouping key ${groupingKey}`);
        }

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
