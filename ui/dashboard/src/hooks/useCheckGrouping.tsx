import BenchmarkNode from "../components/dashboards/check/common/node/BenchmarkNode";
import ControlEmptyResultNode from "../components/dashboards/check/common/node/ControlEmptyResultNode";
import ControlErrorNode from "../components/dashboards/check/common/node/ControlErrorNode";
import ControlNode from "../components/dashboards/check/common/node/ControlNode";
import ControlResultNode from "../components/dashboards/check/common/node/ControlResultNode";
import ControlRunningNode from "../components/dashboards/check/common/node/ControlRunningNode";
import KeyValuePairNode from "../components/dashboards/check/common/node/KeyValuePairNode";
import RootNode from "../components/dashboards/check/common/node/RootNode";
import useCheckFilterConfig from "./useCheckFilterConfig";
import useCheckGroupingConfig from "./useCheckGroupingConfig";
import usePrevious from "./usePrevious";
import {
  AndFilter,
  CheckDisplayGroup,
  CheckNode,
  CheckResult,
  CheckResultDimension,
  CheckResultStatus,
  CheckSeverity,
  CheckSummary,
  CheckTags,
  Filter,
  findDimension,
  OrFilter,
} from "../components/dashboards/check/common";
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from "react";
import { default as BenchmarkType } from "../components/dashboards/check/common/Benchmark";
import { ElementType, IActions, PanelDefinition, PanelsMap } from "../types";
import { useDashboard } from "./useDashboard";
import { useDashboardControls } from "../components/dashboards/layout/Dashboard/DashboardControlsProvider";
import { useSearchParams } from "react-router-dom";

type CheckGroupingActionType = ElementType<typeof checkGroupingActions>;

export type CheckGroupNodeState = {
  expanded: boolean;
};

export type CheckGroupNodeStates = {
  [name: string]: CheckGroupNodeState;
};

export type CheckGroupingAction = {
  type: CheckGroupingActionType;
  [key: string]: any;
};

type CheckGroupFilterStatusValuesMap = {
  [key in keyof typeof CheckResultStatus]: number;
};

export type CheckGroupFilterValues = {
  status: CheckGroupFilterStatusValuesMap;
  dimension: { key: {}; value: {} };
  tag: { key: {}; value: {} };
};

type ICheckGroupingContext = {
  benchmark: BenchmarkType | null;
  definition: PanelDefinition;
  grouping: CheckNode | null;
  groupingsConfig: CheckDisplayGroup[];
  firstChildSummaries: CheckSummary[];
  diffFirstChildSummaries?: CheckSummary[];
  diffGrouping: CheckNode | null;
  nodeStates: CheckGroupNodeStates;
  filterValues: CheckGroupFilterValues;
  dispatch(action: CheckGroupingAction): void;
};

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
  benchmarkChildrenLookup: { [name: string]: CheckNode[] },
): CheckNode => {
  const currentNode = benchmark_trunk.length > 0 ? benchmark_trunk[0] : null;
  let newChildren: CheckNode[];
  if (benchmark_trunk.length > 1) {
    newChildren = [
      addBenchmarkTrunkNode(
        benchmark_trunk.slice(1),
        children,
        benchmarkChildrenLookup,
      ),
    ];
  } else {
    newChildren = children;
  }
  if (!!currentNode?.name) {
    const existingChildren =
      benchmarkChildrenLookup[currentNode?.name || "Other"];
    if (existingChildren) {
      // We only want to add children that are not already in the list,
      // else we end up with duplicate nodes in the tree
      for (const child of newChildren) {
        if (
          existingChildren &&
          existingChildren.find((c) => c.name === child.name)
        ) {
          continue;
        }
        existingChildren.push(child);
      }
    } else {
      benchmarkChildrenLookup[currentNode?.name || "Other"] = newChildren;
    }
  }
  return new BenchmarkNode(
    currentNode?.sort || "Other",
    currentNode?.name || "Other",
    currentNode?.title || "Other",
    newChildren,
  );
};

const getCheckStatusGroupingKey = (status: CheckResultStatus): string => {
  switch (status) {
    case CheckResultStatus.alarm:
      return "Alarm";
    case CheckResultStatus.error:
      return "Error";
    case CheckResultStatus.info:
      return "Info";
    case CheckResultStatus.ok:
      return "OK";
    case CheckResultStatus.skip:
      return "Skipped";
    case CheckResultStatus.empty:
      return "Unknown";
  }
};

const getCheckStatusSortKey = (status: CheckResultStatus): string => {
  switch (status) {
    case CheckResultStatus.ok:
      return "0";
    case CheckResultStatus.alarm:
      return "1";
    case CheckResultStatus.error:
      return "2";
    case CheckResultStatus.info:
      return "3";
    case CheckResultStatus.skip:
      return "4";
    case CheckResultStatus.empty:
      return "5";
  }
};

const getCheckSeverityGroupingKey = (
  severity: CheckSeverity | undefined,
): string => {
  switch (severity) {
    case "critical":
      return "Critical";
    case "high":
      return "High";
    case "low":
      return "Low";
    case "medium":
      return "Medium";
    default:
      return "Unspecified";
  }
};

const getCheckSeveritySortKey = (
  severity: CheckSeverity | undefined,
): string => {
  switch (severity) {
    case "critical":
      return "0";
    case "high":
      return "1";
    case "medium":
      return "2";
    case "low":
      return "3";
    default:
      return "4";
  }
};

const getCheckDimensionGroupingKey = (
  dimensionKey: string | undefined,
  dimensions: CheckResultDimension[],
): string => {
  if (!dimensionKey) {
    return "Dimension key not set";
  }
  const foundDimension = findDimension(dimensions, dimensionKey);
  return foundDimension ? foundDimension.value : `<not set>`;
};

function getCheckTagGroupingKey(tagKey: string | undefined, tags: CheckTags) {
  if (!tagKey) {
    return "Tag key not set";
  }
  return tags[tagKey] || `<not set>`;
}

const getCheckReasonGroupingKey = (reason: string | undefined): string => {
  return reason || "No reason specified.";
};

const getCheckResourceGroupingKey = (resource: string | undefined): string => {
  return resource || "No resource";
};

const getCheckGroupingKey = (
  checkResult: CheckResult,
  group: CheckDisplayGroup,
) => {
  switch (group.type) {
    case "dimension":
      return getCheckDimensionGroupingKey(group.value, checkResult.dimensions);
    case "tag":
      return getCheckTagGroupingKey(group.value, checkResult.tags);
    case "reason":
      return getCheckReasonGroupingKey(checkResult.reason);
    case "resource":
      return getCheckResourceGroupingKey(checkResult.resource);
    case "severity":
      return getCheckSeverityGroupingKey(checkResult.control.severity);
    case "status":
      return getCheckStatusGroupingKey(checkResult.status);
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
  benchmarkChildrenLookup: { [name: string]: CheckNode[] },
): CheckNode => {
  switch (group.type) {
    case "dimension":
      const dimensionValue = getCheckDimensionGroupingKey(
        group.value,
        checkResult.dimensions,
      );
      return new KeyValuePairNode(
        dimensionValue,
        "dimension",
        group.value || "Dimension key not set",
        dimensionValue,
        children,
      );
    case "tag":
      const value = getCheckTagGroupingKey(group.value, checkResult.tags);
      return new KeyValuePairNode(
        value,
        "tag",
        group.value || "Tag key not set",
        value,
        children,
      );
    case "reason":
      return new KeyValuePairNode(
        checkResult.reason || "𤭢", // U+24B62 - very high in sort order - will almost guarantee to put this to the end,
        "reason",
        "reason",
        getCheckReasonGroupingKey(checkResult.reason),
        children,
      );
    case "resource":
      return new KeyValuePairNode(
        checkResult.resource || "𤭢", // U+24B62 - very high in sort order - will almost guarantee to put this to the end
        "resource",
        "resource",
        getCheckResourceGroupingKey(checkResult.resource),
        children,
      );
    case "severity":
      return new KeyValuePairNode(
        getCheckSeveritySortKey(checkResult.control.severity),
        "severity",
        "severity",
        getCheckSeverityGroupingKey(checkResult.control.severity),
        children,
      );
    case "status":
      return new KeyValuePairNode(
        getCheckStatusSortKey(checkResult.status),
        "status",
        "status",
        getCheckStatusGroupingKey(checkResult.status),
        children,
      );
    case "benchmark":
      return checkResult.benchmark_trunk.length > 1
        ? addBenchmarkTrunkNode(
            checkResult.benchmark_trunk.slice(1),
            children,
            benchmarkChildrenLookup,
          )
        : children[0];
    case "control":
      return new ControlNode(
        checkResult.control.sort,
        checkResult.control.name,
        checkResult.control.title,
        children,
      );
    default:
      throw new Error(`Unknown group type ${group.type}`);
  }
};

const addBenchmarkGroupingNode = (
  existingGroups: CheckNode[],
  groupingNode: CheckNode,
) => {
  const existingGroup = existingGroups.find(
    (existingGroup) => existingGroup.name === groupingNode.name,
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
  benchmarkChildrenLookup: { [name: string]: CheckNode[] },
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
            }),
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
          benchmarkChildrenLookup,
        );

        if (groupingNode) {
          if (currentGroupingConfig.type === "benchmark") {
            // For benchmarks, we need to get the benchmark nodes including the trunk
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

type CheckGroupingProviderProps = {
  children: null | JSX.Element | JSX.Element[];
  definition: PanelDefinition;
  diff_panels: PanelsMap | undefined;
};

function recordFilterValues(
  filterValues: {
    severity: { value: {} };
    reason: { value: {} };
    resource: { value: {} };
    control: { value: {} };
    tag: { value: {}; key: {} };
    dimension: { value: {}; key: {} };
    benchmark: { value: {} };
    status: {
      alarm: number;
      skip: number;
      error: number;
      ok: number;
      empty: number;
      info: number;
    };
  },
  checkResult: CheckResult,
) {
  // Record the benchmark of this check result to allow assisted filtering later
  if (!!checkResult.benchmark_trunk && checkResult.benchmark_trunk.length > 0) {
    for (const benchmark of checkResult.benchmark_trunk) {
      filterValues.benchmark.value[benchmark.name] =
        filterValues.benchmark.value[benchmark.name] || 0;
      filterValues.benchmark.value[benchmark.name] += 1;
    }
  }

  // Record the control of this check result to allow assisted filtering later
  filterValues.control.value[checkResult.control.name] =
    filterValues.control.value[checkResult.control.name] || 0;
  filterValues.control.value[checkResult.control.name] += 1;

  // Record the reason of this check result to allow assisted filtering later
  if (checkResult.reason) {
    filterValues.reason.value[checkResult.reason] =
      filterValues.reason.value[checkResult.reason] || 0;
    filterValues.reason.value[checkResult.reason] += 1;
  }

  // Record the resource of this check result to allow assisted filtering later
  if (checkResult.resource) {
    filterValues.resource.value[checkResult.resource] =
      filterValues.resource.value[checkResult.resource] || 0;
    filterValues.resource.value[checkResult.resource] += 1;
  }

  // Record the severity of this check result to allow assisted filtering later
  if (checkResult.severity) {
    filterValues.severity.value[checkResult.severity.toString()] =
      filterValues.severity.value[checkResult.severity.toString()] || 0;
    filterValues.severity.value[checkResult.severity.toString()] += 1;
  }

  // Record the status of this check result to allow assisted filtering later
  if (isNaN(checkResult.status)) {
    filterValues.status[checkResult.status] =
      filterValues.status[checkResult.status] || 0;
    filterValues.status[checkResult.status] += 1;
  }

  // Record the dimension keys/values + value/key counts of this check result to allow assisted filtering later
  for (const dimension of checkResult.dimensions) {
    filterValues.dimension.key[dimension.key] = filterValues.dimension.key[
      dimension.key
      ] || { [dimension.value]: 0 };
    filterValues.dimension.key[dimension.key][dimension.value] += 1;

    filterValues.dimension.value[dimension.value] = filterValues.dimension
      .value[dimension.value] || { [dimension.key]: 0 };
    filterValues.dimension.value[dimension.value][dimension.key] += 1;
  }

  // Record the dimension keys/values + value/key counts of this check result to allow assisted filtering later
  for (const [tagKey, tagValue] of Object.entries(checkResult.tags || {})) {
    filterValues.tag.key[tagKey] = filterValues.tag.key[tagKey] || {
      [tagValue]: 0,
    };
    filterValues.tag.key[tagKey][tagValue] += 1;

    filterValues.tag.value[tagValue] = filterValues.tag.value[tagValue] || {
      [tagKey]: 0,
    };
    filterValues.tag.value[tagValue][tagKey] += 1;
  }
}

const escapeRegex = (string) => {
  if (!string) {
    return string;
  }
  return string.replace(/[-\/\\^$*+?.()|[\]{}]/g, "\\$&");
};

const wildcardToRegex = (wildcard: string) => {
  const escaped = escapeRegex(wildcard);
  return escaped.replaceAll("\\*", ".*");
};

const includeResult = (
  checkResult: CheckResult,
  checkFilterConfig: AndFilter & OrFilter,
): boolean => {
  const andFilter = checkFilterConfig as AndFilter;
  if (!andFilter || !andFilter.and || andFilter.and.length === 0) {
    return true;
  }
  let matches: boolean[] = [];
  for (const filter of andFilter.and) {
    const f = filter as Filter;
    // @ts-ignore
    const valueRegex = new RegExp(`^${wildcardToRegex(f.value)}$`);

    switch (f.type) {
      case "benchmark": {
        let matchesTrunk = false;
        for (const benchmark of checkResult.benchmark_trunk || []) {
          const match = valueRegex.test(benchmark.name);
          if (match) {
            matchesTrunk = true;
            break;
          }
        }
        matches.push(matchesTrunk);
        break;
      }
      case "control": {
        matches.push(valueRegex.test(checkResult.control.name));
        break;
      }
      case "reason": {
        matches.push(valueRegex.test(checkResult.reason));
        break;
      }
      case "resource": {
        matches.push(valueRegex.test(checkResult.resource));
        break;
      }
      case "severity": {
        matches.push(valueRegex.test(checkResult.severity || ""));
        break;
      }
      case "status": {
        matches.push(valueRegex.test(CheckResultStatus[checkResult.status]));
        break;
      }
      case "dimension": {
        // @ts-ignore
        const keyRegex = new RegExp(`^${wildcardToRegex(f.key)}$`);
        let matchesDimensions = false;
        for (const dimension of checkResult.dimensions || []) {
          if (
            keyRegex.test(dimension.key) &&
            valueRegex.test(dimension.value)
          ) {
            matchesDimensions = true;
            break;
          }
        }
        matches.push(matchesDimensions);
        break;
      }
      case "tag": {
        // @ts-ignore
        const keyRegex = new RegExp(`^${wildcardToRegex(f.key)}$`);
        let matchesTags = false;
        for (const [tagKey, tagValue] of Object.entries(
          checkResult.tags || {},
        )) {
          if (keyRegex.test(tagKey) && valueRegex.test(tagValue)) {
            matchesTags = true;
            break;
          }
        }
        matches.push(matchesTags);
        break;
      }
      default:
        matches.push(true);
    }
  }
  return matches.every((m) => m);
};

const useGrouping = (
  definition: PanelDefinition | null,
  panelsMap: PanelsMap | undefined,
  groupingsConfig: CheckDisplayGroup[],
  skip = false,
) => {
  const checkFilterConfig = useCheckFilterConfig();

  return useMemo(() => {
    const filterValues = {
      benchmark: { value: {} },
      control: { value: {} },
      dimension: { key: {}, value: {} },
      reason: { value: {} },
      resource: { value: {} },
      severity: { value: {} },
      status: { alarm: 0, empty: 0, error: 0, info: 0, ok: 0, skip: 0 },
      tag: { key: {}, value: {} },
    };

    if (!definition || skip || !panelsMap) {
      return [null, null, null, [], {}, filterValues];
    }

    // @ts-ignore
    const nestedBenchmarks = definition.children?.filter(
      (child) => child.panel_type === "benchmark",
    );
    const nestedControls =
      definition.panel_type === "control"
        ? [definition]
        : // @ts-ignore
        definition.children?.filter(
          (child) => child.panel_type === "control",
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
      [],
    );

    const checkNodeStates: CheckGroupNodeStates = {};
    const result: CheckNode[] = [];
    const temp = { _: result };
    const benchmarkChildrenLookup = {};

    // We'll loop over each control result and build up the grouped nodes from there
    b.all_control_results.forEach((checkResult) => {
      // See if the result needs to be filtered
      if (!includeResult(checkResult, checkFilterConfig)) {
        return;
      }

      // Build a grouping node - this will be the leaf node down from the root group
      // e.g. benchmark -> control (where control is the leaf)
      const grouping = groupCheckItems(
        temp,
        checkResult,
        groupingsConfig,
        checkNodeStates,
        benchmarkChildrenLookup,
      );
      // Build and add a check result node to the children of the trailing group.
      // This will be used to calculate totals and severity, amongst other things.
      const node = getCheckResultNode(checkResult);
      grouping._.push(node);

      recordFilterValues(filterValues, checkResult);
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
      filterValues,
    ] as const;
  }, [checkFilterConfig, definition, groupingsConfig, panelsMap]);
};

const CheckGroupingProvider = ({
  children,
  definition,
                                 diff_panels,
}: CheckGroupingProviderProps) => {
  const { panelsMap } = useDashboard();
  const { setContext: setDashboardControlsContext } = useDashboardControls();
  const [nodeStates, dispatch] = useReducer(reducer, { nodes: {} });
  const [searchParams] = useSearchParams();
  const groupingsConfig = useCheckGroupingConfig();

  const [
    benchmark,
    panelDefinition,
    grouping,
    firstChildSummaries,
    tempNodeStates,
    filterValues
  ] = useGrouping(definition, panelsMap, groupingsConfig);

  const [, , diffGrouping, diffFirstChildSummaries] = useGrouping(
    definition,
    diff_panels,
    groupingsConfig,
    !diff_panels,
  );

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

  useEffect(() => {
    setDashboardControlsContext(filterValues);
  }, [filterValues, setDashboardControlsContext]);

  return (
    <CheckGroupingContext.Provider
      value={{
        benchmark,
        // @ts-ignore
        definition: panelDefinition,
        dispatch,
        firstChildSummaries,
        diffFirstChildSummaries,
        diffGrouping,
        grouping,
        groupingsConfig,
        nodeStates,
        filterValues,
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
      "useCheckGrouping must be used within a CheckGroupingContext",
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
