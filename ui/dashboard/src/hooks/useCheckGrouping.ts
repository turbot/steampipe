import get from "lodash/get";
import {
  CheckDisplayGroup,
  CheckProps,
} from "../components/dashboards/check/common";
import { default as BenchmarkType } from "../components/dashboards/check/common/Benchmark";
import { useMemo } from "react";

const useCheckGrouping = (props: CheckProps) => {
  const rootBenchmark = get(props, "execution_tree.root.groups[0]", null);

  const groupingsConfig = useMemo(
    () =>
      props.properties?.grouping ||
      ([
        { type: "benchmark" },
        { type: "control" },
        { type: "result" },
      ] as CheckDisplayGroup[]),
    [props.properties]
  );

  return useMemo(() => {
    if (!rootBenchmark) {
      return null;
    }

    const b = new BenchmarkType(
      groupingsConfig,
      0,
      rootBenchmark.group_id,
      rootBenchmark.title,
      rootBenchmark.description,
      rootBenchmark.groups,
      rootBenchmark.controls
    );

    const result: any[] = [];
    const temp = { _: result };
    b.all_control_results.forEach(function (a) {
      [
        { type: "dimension", value: "account_id" },
        { type: "dimension", value: "region" },
      ]
        .reduce(function (r, k) {
          const foundDimension = a.dimensions.find((d) => d.key === k.value);
          const dimension = foundDimension ? foundDimension.value : "None";
          if (!r[dimension]) {
            r[dimension] = { _: [] };
            r._.push({ [k.value]: dimension, ["children"]: r[dimension]._ });
          }
          return r[dimension];
        }, temp)
        ._.push(a);
    });

    return b;
  }, [groupingsConfig, rootBenchmark]);
};

export default useCheckGrouping;
