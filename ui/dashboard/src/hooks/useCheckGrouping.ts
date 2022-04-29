import BenchmarkNode from "../components/dashboards/check/common/node/BenchmarkNode";
import ControlEmptyResultNode from "../components/dashboards/check/common/node/ControlEmptyResultNode";
import ControlErrorNode from "../components/dashboards/check/common/node/ControlErrorNode";
import ControlNode from "../components/dashboards/check/common/node/ControlNode";
import ControlResultNode from "../components/dashboards/check/common/node/ControlResultNode";
import ControlRunningNode from "../components/dashboards/check/common/node/ControlRunningNode";
import get from "lodash/get";
import KeyValuePairNode from "../components/dashboards/check/common/node/KeyValuePairNode";
import RootNode from "../components/dashboards/check/common/node/RootNode";
import {
  CheckDisplayGroup,
  CheckNode,
  CheckProps,
  CheckResult,
  CheckSummary,
  findDimension,
} from "../components/dashboards/check/common";
import { default as BenchmarkType } from "../components/dashboards/check/common/Benchmark";
import { useMemo } from "react";

const addBenchmarkTrunkNode = (
  benchmark_trunk: BenchmarkType[],
  children: CheckNode[]
) => {
  const currentNode = benchmark_trunk.length > 0 ? benchmark_trunk[0] : null;
  return new BenchmarkNode(
    currentNode?.sort || "Other",
    currentNode?.name || "Other",
    currentNode?.title || "Other",
    benchmark_trunk.length > 1
      ? [addBenchmarkTrunkNode(benchmark_trunk.slice(1), children)]
      : children
  );
};

const getCheckGroupingKey = (
  checkResult: CheckResult,
  group: CheckDisplayGroup
) => {
  switch (group.type) {
    case "dimension":
      const foundDimension = findDimension(checkResult.dimensions, group.value);
      return foundDimension ? foundDimension.value : "Other";
    case "tag":
      return group.value ? checkResult.tags[group.value] || "Other" : "Other";
    case "reason":
      return checkResult.reason || "Other";
    case "resource":
      return checkResult.resource || "Other";
    case "severity":
      return checkResult.control.severity || "Other";
    case "status":
      return checkResult.status;
    case "benchmark":
      if (checkResult.benchmark_trunk.length <= 1) {
        return null;
      }
      return checkResult.benchmark_trunk[1].name;
    case "control":
      return checkResult.control.name;
    default:
      return "Other";
  }
};

const getCheckGroupingNode = (
  checkResult: CheckResult,
  group: CheckDisplayGroup,
  children: CheckNode[]
) => {
  switch (group.type) {
    case "dimension":
      const foundDimension = findDimension(checkResult.dimensions, group.value);
      const dimensionValue = foundDimension ? foundDimension.value : "Other";
      return new KeyValuePairNode(
        "dimension",
        group.value || "Other",
        dimensionValue,
        children
      );
    case "tag":
      return new KeyValuePairNode(
        "tag",
        group.value || "Other",
        group.value ? checkResult.tags[group.value] || "Other" : "Other",
        children
      );
    case "reason":
      return new KeyValuePairNode(
        "reason",
        "reason",
        checkResult.reason || "Other",
        children
      );
    case "resource":
      return new KeyValuePairNode(
        "resource",
        "resource",
        checkResult.resource || "Other",
        children
      );
    case "severity":
      return new KeyValuePairNode(
        "severity",
        "severity",
        checkResult.control.severity || "Other",
        children
      );
    case "status":
      return new KeyValuePairNode(
        "status",
        "status",
        checkResult.status,
        children
      );
    case "benchmark":
      return checkResult.benchmark_trunk.length > 1
        ? addBenchmarkTrunkNode(checkResult.benchmark_trunk.slice(1), children)
        : children;
    case "control":
      return new ControlNode(
        checkResult.control.sort,
        checkResult.control.name,
        checkResult.control.title,
        children
      );
    default:
      throw new Error(`Unknown group type ${group.type}`);
  }
};

const groupCheckItems = (
  temp: { _: CheckNode[] },
  group: CheckResult,
  groupingsConfig: CheckDisplayGroup[]
) => {
  return groupingsConfig
    .filter((groupConfig) => groupConfig.type !== "result")
    .reduce(function (grouping, currentGroup) {
      const groupKey = getCheckGroupingKey(group, currentGroup);

      if (!groupKey) {
        return grouping;
      }

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

const getCheckResultNode = (checkResult: CheckResult) => {
  if (checkResult.type === "loading") {
    return new ControlRunningNode(checkResult);
  } else if (checkResult.type === "error") {
    return new ControlErrorNode(checkResult);
  } else if (checkResult.type === "empty") {
    return new ControlEmptyResultNode(checkResult);
  }
  return new ControlResultNode(checkResult);
};

const useCheckGrouping = (props: CheckProps) => {
  const rootBenchmark = get(props, "execution_tree.root.groups[0]", null);

  const groupingsConfig = useMemo(() => {
    if (!rootBenchmark || !rootBenchmark.grouping) {
      return [
        // { type: "status" },
        // { type: "reason" },
        // { type: "resource" },
        // { type: "status" },
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

    return rootBenchmark.grouping;
  }, [rootBenchmark]);

  const [benchmark, grouping, firstChildSummaries] = useMemo(() => {
    if (!rootBenchmark) {
      return [null, null, []];
    }

    const b = new BenchmarkType(
      "0",
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
        getCheckResultNode(checkResult)
      )
    );

    const results = new RootNode(result);

    const firstChildSummaries: CheckSummary[] = [];
    for (const child of results.children) {
      firstChildSummaries.push(child.summary);
    }

    return [b, results, firstChildSummaries] as const;
  }, [groupingsConfig, rootBenchmark]);

  return [
    benchmark,
    grouping,
    groupingsConfig,
    firstChildSummaries,
    rootBenchmark,
  ] as const;
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
