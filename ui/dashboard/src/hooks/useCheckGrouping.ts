import BenchmarkNode from "../components/dashboards/check/common/node/BenchmarkNode";
import Control from "../components/dashboards/check/common/Control";
import ControlNode from "../components/dashboards/check/common/node/ControlNode";
import ControlErrorNode from "../components/dashboards/check/common/node/ControlErrorNode";
import ControlResultNode from "../components/dashboards/check/common/node/ControlResultNode";
import ControlRunningNode from "../components/dashboards/check/common/node/ControlRunningNode";
import get from "lodash/get";
import KeyValuePairNode from "../components/dashboards/check/common/node/KeyValuePairNode";
import padStart from "lodash/padStart";
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
      const root =
        checkResult.benchmark_trunk.length > 1
          ? checkResult.benchmark_trunk[1]
          : null;
      return root ? root.name : "Other";
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
      return addBenchmarkTrunkNode(
        checkResult.benchmark_trunk.length > 1
          ? checkResult.benchmark_trunk.slice(1)
          : [],
        children
      );
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

const addChildren = (node: CheckNode) => {
  const nodes: CheckNode[] = [];
  const lengthMaxIndex = node.children
    ? (node.children?.length - 1).toString().length
    : 0;
  node.children?.forEach((child, index) => {
    if (child.type === "benchmark") {
      nodes.push(
        new BenchmarkNode(
          padStart(index.toString(), lengthMaxIndex),
          child.name,
          child.title,
          addChildren(child)
        )
      );
    } else if (child.type === "control") {
      const control = child as Control;
      const controlChildren: CheckNode[] = [];
      if (control.error) {
        controlChildren.push(
          new ControlErrorNode({
            benchmark_trunk: [],
            dimensions: [],
            reason: "",
            resource: "",
            status: "error",
            tags: {},
            control: {
              name: control.name,
              title: control.title,
              type: "control",
              severity: control.severity,
              severity_summary: {},
              status: "complete",
              sort: "0",
              summary: { error: 1, alarm: 0, ok: 0, info: 0, skip: 0 },
            },
            error: child.error,
          })
        );
      } else if (control.run_state === 4) {
        child.results?.forEach((result) => {
          controlChildren.push(
            new ControlResultNode({
              ...result,
              control: child,
            })
          );
        });
      } else {
        controlChildren.push(new ControlRunningNode(child));
      }
      nodes.push(
        new ControlNode(
          padStart(index.toString(), lengthMaxIndex),
          child.name,
          child.title,
          controlChildren
        )
      );
    }
  });
  return nodes;
};

const buildClassicStructure = (benchmark: BenchmarkType) =>
  new RootNode(addChildren(benchmark));

const useCheckGrouping = (props: CheckProps) => {
  const rootBenchmark = get(props, "execution_tree.root.groups[0]", null);

  // & {
  //     type?: CheckType;
  //     properties: {
  //       display: "all" | "none";
  //       type?: CheckType;
  //       grouping?: CheckDisplayGroup[];
  //     };
  //   }

  const groupingsConfig = useMemo(() => {
    if (!rootBenchmark || !rootBenchmark.grouping) {
      return [
        // { type: "benchmark" },
        // { type: "control" },
        // { type: "result" },
        // { type: "status" },
        // { type: "reason" },
        // { type: "resource" },
        // { type: "severity" },
        // { type: "dimension", value: "account_id" },
        // { type: "status" },
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

    const firstChildSummaries: CheckSummary[] = [];
    for (const child of b.children) {
      firstChildSummaries.push(child.summary);
    }

    const results = buildClassicStructure(b);
    return [b, results, firstChildSummaries] as const;

    // const result: CheckNode[] = [];
    // const temp = { _: result };
    // b.all_control_results.forEach((checkResult) =>
    //   groupCheckItems(temp, checkResult, groupingsConfig)._.push(
    //     new ControlResultNode(checkResult)
    //   )
    // );
    // b.all_control_errors.forEach((checkError) =>
    //   groupCheckItems(temp, checkError, groupingsConfig)._.push(
    //     new ControlErrorNode(checkError)
    //   )
    // );
    // b.all_control_loadings.forEach((checkLoading) =>
    //   groupCheckItems(temp, checkLoading, groupingsConfig)._.push(
    //     new ControlRunningNode(checkLoading.control)
    //   )
    // );
    //
    // const results = new RootNode(result);
    // return [b, results, firstChildSummaries] as const;
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
