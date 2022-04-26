import BenchmarkNode from "../components/dashboards/check/common/BenchmarkNode";
import ControlNode from "../components/dashboards/check/common/ControlNode";
import ControlErrorNode from "../components/dashboards/check/common/ControlErrorNode";
import ControlResultNode from "../components/dashboards/check/common/ControlResultNode";
import get from "lodash/get";
import KeyValuePairNode from "../components/dashboards/check/common/KeyValuePairNode";
import RootNode from "../components/dashboards/check/common/RootNode";
import {
  CheckDisplayGroup,
  CheckNode,
  CheckProps,
  GroupableCheck,
} from "../components/dashboards/check/common";
import { default as BenchmarkType } from "../components/dashboards/check/common/Benchmark";
import { useMemo } from "react";

const addBenchmarkTrunkNode = (
  benchmark_trunk: BenchmarkType[],
  children: CheckNode[]
) => {
  const currentNode = benchmark_trunk.length > 0 ? benchmark_trunk[0] : null;
  return new BenchmarkNode(
    currentNode?.name || "Other",
    currentNode?.title || "Other",
    benchmark_trunk.length > 1
      ? [addBenchmarkTrunkNode(benchmark_trunk.slice(1), children)]
      : children
  );
};

const getCheckGroupingKey = (
  check: GroupableCheck,
  group: CheckDisplayGroup
) => {
  switch (group.type) {
    case "dimension":
      const foundDimension = check.dimensions.find(
        (d) => d.key === group.value
      );
      return foundDimension ? foundDimension.value : "Other";
    case "tag":
      return group.value ? check.tags[group.value] || "Other" : "Other";
    case "benchmark":
      const root =
        check.benchmark_trunk.length > 1 ? check.benchmark_trunk[1] : null;
      return root ? root.name : "Other";
    case "control":
      return check.control.name;
    default:
      return "Other";
  }
};

const getCheckGroupingNode = (
  check: GroupableCheck,
  group: CheckDisplayGroup,
  children: CheckNode[]
) => {
  switch (group.type) {
    case "dimension":
      const foundDimension = check.dimensions.find(
        (d) => d.key === group.value
      );
      const dimensionValue = foundDimension ? foundDimension.value : "Other";
      return new KeyValuePairNode(
        group.value || "Other",
        dimensionValue,
        children
      );
    case "tag":
      return new KeyValuePairNode(
        group.value || "Other",
        group.value ? check.tags[group.value] || "Other" : "Other",
        children
      );
    case "benchmark":
      return addBenchmarkTrunkNode(
        check.benchmark_trunk.length > 1 ? check.benchmark_trunk.slice(1) : [],
        children
      );
    case "control":
      return new ControlNode(check.control.name, check.control.title, children);
    default:
      throw new Error(`Unknown group type ${group.type}`);
  }
};

const groupCheckItems = (
  temp: { _: CheckNode[] },
  group: GroupableCheck,
  groupingsConfig: CheckDisplayGroup[]
) => {
  return groupingsConfig
    .filter((groupConfig) => groupConfig.type !== "result")
    .reduce(function (grouping, currentGroup) {
      const groupKey = getCheckGroupingKey(group, currentGroup);
      if (!grouping[groupKey]) {
        grouping[groupKey] = { _: [] };
        const groupingNode = getCheckGroupingNode(
          group,
          currentGroup,
          grouping[groupKey]._
        );

        if (groupingNode) {
          grouping._.push(groupingNode);
        }
      }

      return grouping[groupKey];
    }, temp);
};

const useCheckGrouping = (props: CheckProps) => {
  const rootBenchmark = get(props, "execution_tree.root.groups[0]", null);

  const groupingsConfig = useMemo(
    () =>
      props.properties?.grouping ||
      ([
        // { type: "benchmark" },
        // { type: "control" },
        // { type: "result" },
        // { type: "dimension", value: "account_id" },
        { type: "dimension", value: "region" },
        // { type: "dimension", value: "region" },
        { type: "tag", value: "service" },
        // { type: "tag", value: "cis_type" },
        // { type: "benchmark" },
        // { type: "control" },
        { type: "result" },
      ] as CheckDisplayGroup[]),
    [props.properties]
  );

  const grouping = useMemo(() => {
    if (!rootBenchmark) {
      return null;
    }

    const b = new BenchmarkType(
      rootBenchmark.group_id,
      rootBenchmark.title,
      rootBenchmark.description,
      rootBenchmark.groups,
      rootBenchmark.controls,
      []
    );

    const result: CheckNode[] = [];
    const temp = { _: result };
    b.all_control_results.forEach((checkResult) =>
      groupCheckItems(temp, checkResult, groupingsConfig)._.push(
        new ControlResultNode(checkResult)
      )
    );
    b.all_control_errors.forEach((checkError) =>
      groupCheckItems(temp, checkError, groupingsConfig)._.push(
        new ControlErrorNode(checkError)
      )
    );
    // b.all_control_errors.forEach(function (checkError) {
    //   return groupingsConfig
    //     .filter((group) => group.type !== "result")
    //     .reduce(function (grouping, currentGroup) {
    //       const groupKey = getCheckGroupingKey(checkError, currentGroup);
    //       if (!grouping[groupKey]) {
    //         grouping[groupKey] = { _: [] };
    //         const groupingNode = getCheckGroupingNode(
    //           checkError,
    //           currentGroup,
    //           grouping[groupKey]._
    //         );
    //
    //         if (groupingNode) {
    //           grouping._.push(groupingNode);
    //         }
    //       }
    //
    //       return grouping[groupKey];
    //     }, temp)
    //     ._.push(new ControlErrorNode(checkError));
    // });

    return new RootNode(result);
  }, [groupingsConfig, rootBenchmark]);

  return [grouping, groupingsConfig] as const;
};

export default useCheckGrouping;

// https://stackoverflow.com/questions/50737098/multi-level-grouping-in-javascript
// keys = ['level1', 'level2'],
//     result = [],
//     temp = { _: result };
//
// data.forEach(function (a) {
//   keys.reduce(function (r, k) {
//     if (!r[a[k]]) {
//       r[a[k]] = { _: [] };
//       r._.push({ [k]: a[k], [k + 'list']: r[a[k]]._ });
//     }
//     return r[a[k]];
//   }, temp)._.push({ Id: a.Id });
// });
//
// console.log(result);
