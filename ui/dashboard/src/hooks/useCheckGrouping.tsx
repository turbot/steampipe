import BenchmarkNode from "../components/dashboards/check/common/node/BenchmarkNode";
import ControlEmptyResultNode from "../components/dashboards/check/common/node/ControlEmptyResultNode";
import ControlErrorNode from "../components/dashboards/check/common/node/ControlErrorNode";
import ControlNode from "../components/dashboards/check/common/node/ControlNode";
import ControlResultNode from "../components/dashboards/check/common/node/ControlResultNode";
import ControlRunningNode from "../components/dashboards/check/common/node/ControlRunningNode";
import KeyValuePairNode from "../components/dashboards/check/common/node/KeyValuePairNode";
import RootNode from "../components/dashboards/check/common/node/RootNode";
import usePrevious from "./usePrevious";
import {
  CheckDisplayGroup,
  CheckDisplayGroupType,
  CheckNode,
  CheckResult,
  CheckSummary,
  findDimension,
} from "../components/dashboards/check/common";
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from "react";
import { default as BenchmarkType } from "../components/dashboards/check/common/Benchmark";
import { ElementType, IActions, PanelDefinition } from "../types";
import { useDashboard } from "./useDashboard";
import { useSearchParams } from "react-router-dom";

type CheckGroupingActionType = ElementType<typeof checkGroupingActions>;

export interface CheckGroupNodeState {
  expanded: boolean;
}

export interface CheckGroupNodeStates {
  [name: string]: CheckGroupNodeState;
}

export interface CheckGroupingAction {
  type: CheckGroupingActionType;
  [key: string]: any;
}

interface ICheckGroupingContext {
  benchmark: BenchmarkType | null;
  definition: PanelDefinition;
  grouping: CheckNode | null;
  groupingsConfig: CheckDisplayGroup[];
  firstChildSummaries: CheckSummary[];
  nodeStates: CheckGroupNodeStates;
  dispatch(action: CheckGroupingAction): void;
}

const CheckGroupingActions: IActions = {
  COLLAPSE_ALL_NODES: "collapse_all_nodes",
  COLLAPSE_NODE: "collapse_node",
  EXPAND_ALL_NODES: "expand_all_nodes",
  EXPAND_NODE: "expand_node",
  UPDATE_NODES: "update_nodes",
};

const checkGroupingActions = Object.values(CheckGroupingActions);

const CheckGroupingContext = createContext<ICheckGroupingContext | null>(null);

const addBenchmarkTrunkNode = (
  benchmark_trunk: BenchmarkType[],
  children: CheckNode[],
  benchmarkChildrenLookup: { [name: string]: CheckNode[] }
): CheckNode => {
  const currentNode = benchmark_trunk.length > 0 ? benchmark_trunk[0] : null;
  let newChildren: CheckNode[];
  if (benchmark_trunk.length > 1) {
    newChildren = [
      addBenchmarkTrunkNode(
        benchmark_trunk.slice(1),
        children,
        benchmarkChildrenLookup
      ),
    ];
  } else {
    newChildren = children;
  }
  if (!!currentNode?.name) {
    const existingChildren =
      benchmarkChildrenLookup[currentNode?.name || "Other"];
    if (existingChildren) {
      existingChildren.push(...newChildren);
    } else {
      benchmarkChildrenLookup[currentNode?.name || "Other"] = newChildren;
    }
  }
  return new BenchmarkNode(
    currentNode?.sort || "Other",
    currentNode?.name || "Other",
    currentNode?.title || "Other",
    newChildren
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
      return checkResult.status === "empty" ? "Other" : checkResult.status;
    case "benchmark":
      if (checkResult.benchmark_trunk.length <= 1) {
        return null;
      }
      return checkResult.benchmark_trunk[checkResult.benchmark_trunk.length - 1]
        .name;
    case "control":
      return checkResult.control.name;
    default:
      return "Other";
  }
};

const getCheckGroupingNode = (
  checkResult: CheckResult,
  group: CheckDisplayGroup,
  children: CheckNode[],
  benchmarkChildrenLookup: { [name: string]: CheckNode[] }
): CheckNode => {
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
        checkResult.status === "empty" ? "Other" : checkResult.status,
        children
      );
    case "benchmark":
      return checkResult.benchmark_trunk.length > 1
        ? addBenchmarkTrunkNode(
            checkResult.benchmark_trunk.slice(1),
            children,
            benchmarkChildrenLookup
          )
        : children[0];
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

const addBenchmarkGroupingNode = (
  existingGroups: CheckNode[],
  groupingNode: CheckNode
) => {
  const existingGroup = existingGroups.find(
    (existingGroup) => existingGroup.name === groupingNode.name
  );
  if (existingGroup) {
    (existingGroup as BenchmarkNode).merge(groupingNode);
  } else {
    existingGroups.push(groupingNode);
  }
};

const groupCheckItems = (
  temp: { _: CheckNode[] },
  checkResult: CheckResult,
  groupingsConfig: CheckDisplayGroup[],
  checkNodeStates: CheckGroupNodeStates,
  benchmarkChildrenLookup: { [name: string]: CheckNode[] }
) => {
  return groupingsConfig
    .filter((groupConfig) => groupConfig.type !== "result")
    .reduce((cumulativeGrouping, currentGroupingConfig) => {
      // Get this items grouping key - e.g. control or benchmark name
      const groupKey = getCheckGroupingKey(checkResult, currentGroupingConfig);

      if (!groupKey) {
        return cumulativeGrouping;
      }

      // Collapse all benchmark trunk nodes
      if (currentGroupingConfig.type === "benchmark") {
        checkResult.benchmark_trunk.forEach(
          (benchmark) =>
            (checkNodeStates[benchmark.name] = {
              expanded: false,
            })
        );
      } else {
        checkNodeStates[groupKey] = {
          expanded: false,
        };
      }

      if (!cumulativeGrouping[groupKey]) {
        cumulativeGrouping[groupKey] = { _: [] };

        const groupingNode = getCheckGroupingNode(
          checkResult,
          currentGroupingConfig,
          cumulativeGrouping[groupKey]._,
          benchmarkChildrenLookup
        );

        if (groupingNode) {
          if (currentGroupingConfig.type === "benchmark") {
            // For benchmarks we need to get the benchmark nodes including the trunk
            addBenchmarkGroupingNode(cumulativeGrouping._, groupingNode);
          } else {
            cumulativeGrouping._.push(groupingNode);
          }
        }
      }

      // If the grouping key for this has already been logged by another result,
      // use the existing children from that - this covers cases where we may have
      // benchmark 1 -> benchmark 2 -> control 1
      // benchmark 1 -> control 2
      // ...when we build the benchmark grouping node for control 1, its key will be
      // for benchmark 2, but we'll add a hierarchical grouping node for benchmark 1 -> benchmark 2
      // When we come to get the benchmark grouping node for control 2, we'll need to add
      // the control to the existing children of benchmark 1
      if (
        currentGroupingConfig.type === "benchmark" &&
        benchmarkChildrenLookup[groupKey]
      ) {
        const groupingEntry = cumulativeGrouping[groupKey];
        const { _, ...rest } = groupingEntry || {};
        cumulativeGrouping[groupKey] = {
          _: benchmarkChildrenLookup[groupKey],
          ...rest,
        };
      }

      return cumulativeGrouping[groupKey];
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

const reducer = (state: CheckGroupNodeStates, action) => {
  switch (action.type) {
    case CheckGroupingActions.COLLAPSE_ALL_NODES: {
      const newNodes = {};
      for (const [name, node] of Object.entries(state)) {
        newNodes[name] = {
          ...node,
          expanded: false,
        };
      }
      return {
        ...state,
        nodes: newNodes,
      };
    }
    case CheckGroupingActions.COLLAPSE_NODE:
      return {
        ...state,
        [action.name]: {
          ...(state[action.name] || {}),
          expanded: false,
        },
      };
    case CheckGroupingActions.EXPAND_ALL_NODES: {
      const newNodes = {};
      Object.entries(state).forEach(([name, node]) => {
        newNodes[name] = {
          ...node,
          expanded: true,
        };
      });
      return newNodes;
    }
    case CheckGroupingActions.EXPAND_NODE: {
      return {
        ...state,
        [action.name]: {
          ...(state[action.name] || {}),
          expanded: true,
        },
      };
    }
    case CheckGroupingActions.UPDATE_NODES:
      return action.nodes;
    default:
      return state;
  }
};

interface CheckGroupingProviderProps {
  children: null | JSX.Element | JSX.Element[];
  definition: PanelDefinition;
}

const CheckGroupingProvider = ({
  children,
  definition,
}: CheckGroupingProviderProps) => {
  const { panelsMap } = useDashboard();
  const [nodeStates, dispatch] = useReducer(reducer, { nodes: {} });
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
        { type: "benchmark" },
        { type: "control" },
        { type: "result" },
      ] as CheckDisplayGroup[];
    }
  }, [searchParams]);

  const [
    benchmark,
    panelDefinition,
    grouping,
    firstChildSummaries,
    tempNodeStates,
  ] = useMemo(() => {
    if (!definition) {
      return [null, null, null, [], {}];
    }

    // @ts-ignore
    const nestedBenchmarks = definition.children?.filter(
      (child) => child.panel_type === "benchmark"
    );
    const nestedControls =
      definition.panel_type === "control"
        ? [definition]
        : // @ts-ignore
          definition.children?.filter(
            (child) => child.panel_type === "control"
          );

    const rootBenchmarkPanel = panelsMap[definition.name];
    const b = new BenchmarkType(
      "0",
      rootBenchmarkPanel.name,
      rootBenchmarkPanel.title,
      rootBenchmarkPanel.description,
      nestedBenchmarks,
      nestedControls,
      panelsMap,
      []
    );

    const checkNodeStates: CheckGroupNodeStates = {};
    const result: CheckNode[] = [];
    const temp = { _: result };
    const benchmarkChildrenLookup = {};

    // We'll loop over each control result and build up the grouped nodes from there
    b.all_control_results.forEach((checkResult) => {
      // Build a grouping node - this will be the leaf node down from the root group
      // e.g. benchmark -> control (where control is the leaf)
      const grouping = groupCheckItems(
        temp,
        checkResult,
        groupingsConfig,
        checkNodeStates,
        benchmarkChildrenLookup
      );
      // Build and add a check result node to the children of the trailing group.
      // This will be used to calculate totals and severity, amongst other things.
      const node = getCheckResultNode(checkResult);
      grouping._.push(node);
    });

    const results = new RootNode(result);

    const firstChildSummaries: CheckSummary[] = [];
    for (const child of results.children) {
      firstChildSummaries.push(child.summary);
    }

    return [
      b,
      { ...rootBenchmarkPanel, children: definition.children },
      results,
      firstChildSummaries,
      checkNodeStates,
    ] as const;
  }, [definition, groupingsConfig, panelsMap]);

  const previousGroupings = usePrevious({ groupingsConfig });

  useEffect(() => {
    if (
      previousGroupings &&
      // @ts-ignore
      previousGroupings.groupingsConfig === groupingsConfig
    ) {
      return;
    }
    dispatch({
      type: CheckGroupingActions.UPDATE_NODES,
      nodes: tempNodeStates,
    });
  }, [previousGroupings, groupingsConfig, tempNodeStates]);

  return (
    <CheckGroupingContext.Provider
      value={{
        benchmark,
        // @ts-ignore
        definition: panelDefinition,
        dispatch,
        firstChildSummaries,
        grouping,
        groupingsConfig,
        nodeStates,
      }}
    >
      {children}
    </CheckGroupingContext.Provider>
  );
};

const useCheckGrouping = () => {
  const context = useContext(CheckGroupingContext);
  if (context === undefined) {
    throw new Error(
      "useCheckGrouping must be used within a CheckGroupingContext"
    );
  }
  return context as ICheckGroupingContext;
};

export {
  CheckGroupingActions,
  CheckGroupingContext,
  CheckGroupingProvider,
  useCheckGrouping,
};

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
