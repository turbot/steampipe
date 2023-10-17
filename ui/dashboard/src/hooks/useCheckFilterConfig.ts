import {
  CheckFilter,
  CheckFilterType,
} from "../components/dashboards/check/common";
import { useMemo } from "react";
import { useSearchParams } from "react-router-dom";

const filterKeys = [
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

const useCheckFilterConfig = () => {
  const [searchParams] = useSearchParams();
  return useMemo(() => {
    const rawFilters = searchParams.get("filter");
    if (rawFilters) {
      const filters: CheckFilter[] = [];
      const filterParts = rawFilters.split(",").filter((g) => !!g);
      for (const groupingPart of filterParts) {
        const typeValueParts = groupingPart.split("|");
        const groupingKey = typeValueParts[0];

        // Is this a valid grouping key?
        const isValid = filterKeys.includes(groupingKey);
        if (!isValid) {
          throw new Error(`Unsupported grouping key ${groupingKey}`);
        }

        if (typeValueParts.length > 1) {
          filters.push({
            id: `${typeValueParts[0]}-${typeValueParts[1]}`,
            type: typeValueParts[0] as CheckFilterType,
            value: typeValueParts[1],
          });
        } else {
          filters.push({
            id: typeValueParts[0],
            type: typeValueParts[0] as CheckFilterType,
          });
        }
      }
      return filters;
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
      ] as CheckFilter[];
    }
  }, [searchParams]);
};

export default useCheckFilterConfig;
