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
  CheckBenchmarkRun,
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
import { ElementType, IActions, PanelDefinition } from "./useDashboard";
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
  rootBenchmark: CheckBenchmarkRun;
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
  children: CheckNode[]
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
  checkNodeStates: CheckGroupNodeStates
) => {
  return groupingsConfig
    .filter((groupConfig) => groupConfig.type !== "result")
    .reduce(function (cumulativeGrouping, currentGroupingConfig) {
      const groupKey = getCheckGroupingKey(checkResult, currentGroupingConfig);

      if (!groupKey) {
        return cumulativeGrouping;
      }

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
          cumulativeGrouping[groupKey]._
        );

        if (groupingNode) {
          if (currentGroupingConfig.type === "benchmark") {
            addBenchmarkGroupingNode(cumulativeGrouping._, groupingNode);
          } else {
            cumulativeGrouping._.push(groupingNode);
          }
        }
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
  const [nodeStates, dispatch] = useReducer(reducer, { nodes: {} });
  const rootBenchmark = get(
    definition,
    "execution_tree.root.children[0]",
    null
  );
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
  }, [searchParams]);

  const [benchmark, grouping, firstChildSummaries, tempNodeStates] =
    useMemo(() => {
      if (!rootBenchmark) {
        return [null, null, [], {}];
      }

      // @ts-ignore
      const nestedBenchmarks = rootBenchmark.children?.filter(
        (child) => child.node_type === "benchmark_run"
      );
      // @ts-ignore
      const nestedControls = rootBenchmark.children?.filter(
        (child) => child.node_type === "control_run"
      );

      const b = new BenchmarkType(
        "0",
        rootBenchmark.name,
        rootBenchmark.title,
        rootBenchmark.description,
        nestedBenchmarks,
        nestedControls,
        []
      );

      const checkNodeStates: CheckGroupNodeStates = {};
      const result: CheckNode[] = [];
      const temp = { _: result };
      b.all_control_results.forEach((checkResult) => {
        const grouping = groupCheckItems(
          temp,
          checkResult,
          groupingsConfig,
          checkNodeStates
        );
        const node = getCheckResultNode(checkResult);
        grouping._.push(node);
      });

      const results = new RootNode(result);

      const firstChildSummaries: CheckSummary[] = [];
      for (const child of results.children) {
        firstChildSummaries.push(child.summary);
      }

      return [b, results, firstChildSummaries, checkNodeStates] as const;
    }, [groupingsConfig, rootBenchmark]);

  useEffect(() => {
    dispatch({
      type: CheckGroupingActions.UPDATE_NODES,
      nodes: tempNodeStates,
    });
  }, [groupingsConfig]);

  return (
    <CheckGroupingContext.Provider
      value={{
        benchmark,
        definition,
        dispatch,
        firstChildSummaries,
        grouping,
        groupingsConfig,
        nodeStates,
        rootBenchmark,
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
