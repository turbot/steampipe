import findPathDeep from "deepdash/findPathDeep";
import get from "lodash/get";
import isEqual from "lodash/isEqual";
import paths from "deepdash/paths";
import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useReducer,
  useState,
} from "react";
import set from "lodash/set";
import sortBy from "lodash/sortBy";
import useDashboardWebSocket, { SocketActions } from "./useDashboardWebSocket";
import usePrevious from "./usePrevious";
import { buildComponentsMap } from "../components";
import { CheckExecutionTree } from "../components/dashboards/check/common";
import { GlobalHotKeys } from "react-hotkeys";
import { LeafNodeData, Width } from "../components/dashboards/common";
import { noop } from "../utils/func";
import { Theme } from "./useTheme";
import {
  useLocation,
  useNavigate,
  useNavigationType,
  useParams,
  useSearchParams,
} from "react-router-dom";

interface IBreakpointContext {
  currentBreakpoint: string | null;
  maxBreakpoint(breakpointAndDown: string): boolean;
  minBreakpoint(breakpointAndUp: string): boolean;
  width: number;
}

interface IThemeContext {
  theme: Theme;
  setTheme(theme: string): void;
  wrapperRef: React.Ref<null>;
}

export interface ComponentsMap {
  [name: string]: any;
}

export type DashboardDataMode = "live" | "snapshot";

interface IDashboardContext {
  metadata: DashboardMetadata | null;
  availableDashboardsLoaded: boolean;

  closePanelDetail(): void;
  dispatch(action: DashboardAction): void;

  dataMode: DashboardDataMode;
  refetchDashboard: boolean;

  error: any;

  dashboards: AvailableDashboard[];
  dashboardsMap: AvailableDashboardsDictionary;
  dashboard: DashboardDefinition | null;

  selectedPanel: PanelDefinition | null;
  selectedDashboard: AvailableDashboard | null;
  selectedDashboardInputs: DashboardInputs;
  selectedSnapshot: DashboardSnapshot | null;
  lastChangedInput: string | null;

  sqlDataMap: SQLDataMap;

  dashboardTags: DashboardTags;

  search: DashboardSearch;

  breakpointContext: IBreakpointContext;
  themeContext: IThemeContext;

  components: ComponentsMap;
}

export interface IActions {
  [type: string]: string;
}

const DashboardActions: IActions = {
  AVAILABLE_DASHBOARDS: "available_dashboards",
  CLEAR_DASHBOARD_INPUTS: "clear_dashboard_inputs",
  CLEAR_SNAPSHOT: "clear_snapshot",
  CONTROL_COMPLETE: "control_complete",
  CONTROL_ERROR: "control_error",
  DASHBOARD_METADATA: "dashboard_metadata",
  DELETE_DASHBOARD_INPUT: "delete_dashboard_input",
  EXECUTION_COMPLETE: "execution_complete",
  EXECUTION_ERROR: "execution_error",
  EXECUTION_STARTED: "execution_started",
  INPUT_VALUES_CLEARED: "input_values_cleared",
  LEAF_NODE_COMPLETE: "leaf_node_complete",
  LEAF_NODE_PROGRESS: "leaf_node_progress",
  SELECT_DASHBOARD: "select_dashboard",
  SELECT_PANEL: "select_panel",
  SELECT_SNAPSHOT: "select_snapshot",
  SET_DASHBOARD: "set_dashboard",
  SET_DASHBOARD_INPUT: "set_dashboard_input",
  SET_DASHBOARD_INPUTS: "set_dashboard_inputs",
  SET_DASHBOARD_SEARCH_VALUE: "set_dashboard_search_value",
  SET_DASHBOARD_SEARCH_GROUP_BY: "set_dashboard_search_group_by",
  SET_DASHBOARD_TAG_KEYS: "set_dashboard_tag_keys",
  SET_DATA_MODE: "set_data_mode",
  SET_REFETCH_DASHBOARD: "set_refetch_dashboard",
  SET_SNAPSHOT: "set_snapshot",
  WORKSPACE_ERROR: "workspace_error",
};

const dashboardActions = Object.values(DashboardActions);

// https://github.com/microsoft/TypeScript/issues/28046
export type ElementType<T extends ReadonlyArray<unknown>> =
  T extends ReadonlyArray<infer ElementType> ? ElementType : never;

export type DashboardActionType = ElementType<typeof dashboardActions>;

export interface DashboardAction {
  type: DashboardActionType;
  [key: string]: any;
}

type DashboardSearchGroupByMode = "mod" | "tag";

interface DashboardSearchGroupBy {
  value: DashboardSearchGroupByMode;
  tag: string | null;
}

export interface DashboardSearch {
  value: string;
  groupBy: DashboardSearchGroupBy;
}

export interface DashboardTags {
  keys: string[];
}

interface SelectedDashboardStates {
  dashboard_name: string | null;
  dataMode: DashboardDataMode;
  refetchDashboard: boolean;
  search: DashboardSearch;
  selectedDashboard: AvailableDashboard | null;
  selectedDashboardInputs: DashboardInputs;
  selectedSnapshot: DashboardSnapshot;
}

interface DashboardInputs {
  [name: string]: string;
}

interface DashboardVariables {
  [name: string]: any;
}

export interface ModDashboardMetadata {
  title: string;
  full_name: string;
  short_name: string;
}

interface InstalledModsDashboardMetadata {
  [key: string]: ModDashboardMetadata;
}

export interface CloudDashboardActorMetadata {
  id: string;
  handle: string;
}

export interface CloudDashboardIdentityMetadata {
  id: string;
  handle: string;
  type: "org" | "user";
}

export interface CloudDashboardWorkspaceMetadata {
  id: string;
  handle: string;
}

interface CloudDashboardMetadata {
  actor: CloudDashboardActorMetadata;
  identity: CloudDashboardIdentityMetadata;
  workspace: CloudDashboardWorkspaceMetadata;
}

export interface DashboardMetadata {
  mod: ModDashboardMetadata;
  installed_mods?: InstalledModsDashboardMetadata;
  cloud?: CloudDashboardMetadata;
  telemetry: "info" | "none";
}

export interface DashboardSnapshot {
  id: string;
  dashboard_name: string;
  start_time: string;
  end_time: string;
  lineage: string;
  schema_version: string;
  search_path: string;
  variables: DashboardVariables;
  inputs: DashboardInputs;
}

interface AvailableDashboardTags {
  [key: string]: string;
}

type AvailableDashboardType = "benchmark" | "dashboard";

export interface AvailableDashboard {
  full_name: string;
  short_name: string;
  mod_full_name: string;
  tags: AvailableDashboardTags;
  title: string;
  is_top_level: boolean;
  type: AvailableDashboardType;
  children?: AvailableDashboard[];
  trunks?: string[][];
}

export interface AvailableDashboardsDictionary {
  [key: string]: AvailableDashboard;
}

export interface ContainerDefinition {
  name: string;
  node_type?: string;
  allow_child_panel_expand?: boolean;
  data?: LeafNodeData;
  title?: string;
  width?: number;
  children?: (ContainerDefinition | PanelDefinition)[];
}

export interface PanelProperties {
  [key: string]: any;
}

export interface SQLDataMap {
  [sql: string]: LeafNodeData;
}

export interface PanelDefinition {
  name: string;
  node_type?: string;
  title?: string;
  width?: Width;
  sql?: string;
  data?: LeafNodeData;
  source_definition?: string;
  execution_tree?: CheckExecutionTree;
  error?: Error;
  properties?: PanelProperties;
  dashboard: string;
}

export interface DashboardDefinition {
  artificial: boolean;
  name: string;
  node_type: string;
  title?: string;
  width?: number;
  children?: (ContainerDefinition | PanelDefinition)[];
  dashboard: string;
}

interface DashboardsCollection {
  dashboards: AvailableDashboard[];
  dashboardsMap: AvailableDashboardsDictionary;
}

const buildDashboards = (
  dashboards: AvailableDashboardsDictionary,
  benchmarks: AvailableDashboardsDictionary
): DashboardsCollection => {
  const dashboardsMap = {};
  const builtDashboards: AvailableDashboard[] = [];

  for (const [, dashboard] of Object.entries(dashboards)) {
    const builtDashboard: AvailableDashboard = {
      title: dashboard.title,
      full_name: dashboard.full_name,
      short_name: dashboard.short_name,
      type: "dashboard",
      tags: dashboard.tags,
      mod_full_name: dashboard.mod_full_name,
      is_top_level: true,
    };
    dashboardsMap[builtDashboard.full_name] = builtDashboard;
    builtDashboards.push(builtDashboard);
  }

  for (const [, benchmark] of Object.entries(benchmarks)) {
    const builtBenchmark: AvailableDashboard = {
      title: benchmark.title,
      full_name: benchmark.full_name,
      short_name: benchmark.short_name,
      type: "benchmark",
      tags: benchmark.tags,
      mod_full_name: benchmark.mod_full_name,
      is_top_level: benchmark.is_top_level,
      trunks: benchmark.trunks,
      children: benchmark.children,
    };
    dashboardsMap[builtBenchmark.full_name] = builtBenchmark;
    builtDashboards.push(builtBenchmark);
  }

  return {
    dashboards: sortBy(builtDashboards, [
      (dashboard) =>
        dashboard.title
          ? dashboard.title.toLowerCase()
          : dashboard.full_name.toLowerCase(),
    ]),
    dashboardsMap,
  };
};

const updateSelectedDashboard = (
  selectedDashboard: AvailableDashboard,
  newDashboards: AvailableDashboard[]
) => {
  if (!selectedDashboard) {
    return null;
  }
  const matchingDashboard = newDashboards.find(
    (dashboard) => dashboard.full_name === selectedDashboard.full_name
  );
  if (matchingDashboard) {
    return matchingDashboard;
  } else {
    return null;
  }
};

function buildSqlDataMap(dashboard: DashboardDefinition): SQLDataMap {
  const sqlPaths = paths(dashboard, { leavesOnly: true }).filter((path) =>
    path.endsWith(".sql")
  );
  const sqlDataMap = {};
  for (const sqlPath of sqlPaths) {
    const sql = get(dashboard, sqlPath);
    const dataPath = `${sqlPath.substring(0, sqlPath.indexOf(".sql"))}.data`;
    const data = get(dashboard, dataPath);
    if (!sqlDataMap[sql]) {
      sqlDataMap[sql] = data;
    }
  }
  return sqlDataMap;
}

function addDataToDashboard(
  dashboard: DashboardDefinition,
  sqlDataMap: SQLDataMap
): DashboardDefinition {
  const sqlPaths = paths(dashboard, { leavesOnly: true }).filter((path) =>
    path.endsWith(".sql")
  );
  for (const sqlPath of sqlPaths) {
    const sql = get(dashboard, sqlPath);
    const data = sqlDataMap[sql];
    if (!data) {
      continue;
    }
    const dataPath = `${sqlPath.substring(0, sqlPath.indexOf(".sql"))}.data`;
    set(dashboard, dataPath, data);
  }
  return dashboard;
}

const wrapDefinitionInArtificialDashboard = (
  definition: DashboardDefinition
): DashboardDefinition => {
  const { title, ...definitionWithoutTitle } = definition;
  return {
    artificial: true,
    name: definition.name,
    title: definition.title,
    node_type: "dashboard",
    children: [
      {
        ...definitionWithoutTitle,
      },
    ],
    dashboard: definition.dashboard,
  };
};

const updateCheckNode = (dashboardCheckNode, action) => {
  let panelPath: string = findPathDeep(
    dashboardCheckNode,
    (v, k) => k === "control_id" && v === action.control.control_id
  );

  if (!panelPath) {
    console.warn("Cannot find control to update", action.control.control_id);
    return null;
  }

  panelPath = panelPath.replace(".control_id", "");

  let newCheckNode = {
    ...dashboardCheckNode,
  };

  return set(newCheckNode, panelPath, action.control);
};

const updateDashboardWithControlEvent = (dashboard, action) => {
  if (dashboard.artificial) {
    let updatedCheckNode = updateCheckNode(
      get(dashboard, "children[0]"),
      action
    );

    if (!updatedCheckNode) {
      console.warn("Cannot find control to update", action.control.control_id);
      return null;
    }

    const rootBenchmark = get(
      updatedCheckNode,
      "execution_tree.root.groups[0]",
      {}
    );

    updatedCheckNode = set(updatedCheckNode, "execution_tree.root.groups[0]", {
      ...rootBenchmark,
    });

    const newDashboard = {
      ...dashboard,
    };

    return set(newDashboard, "children[0]", updatedCheckNode);
  } else {
    let nodePath: string = findPathDeep(
      dashboard,
      (v, k) => k === "name" && v === action.name
    );

    if (!nodePath) {
      console.warn("Cannot find dashboard node to update", action.name);
      return null;
    }

    nodePath = nodePath.replace(".name", "");

    let node = get(dashboard, nodePath);

    const rootBenchmark = get(node, "execution_tree.root.groups[0]", {});
    node = set(node, "execution_tree.root.groups[0]", { ...rootBenchmark });

    if (!node) {
      console.warn("Cannot find dashboard node to update", action.name);
      return null;
    }

    let updatedNode = updateCheckNode(node, action);

    if (!updatedNode) {
      console.warn("Cannot find control to update", action.control.control_id);
      return null;
    }

    return set(dashboard, nodePath, updatedNode);
  }
};

function reducer(state, action) {
  switch (action.type) {
    case DashboardActions.DASHBOARD_METADATA:
      return {
        ...state,
        metadata: action.metadata,
      };
    case DashboardActions.AVAILABLE_DASHBOARDS:
      const { dashboards, dashboardsMap } = buildDashboards(
        action.dashboards,
        action.benchmarks
      );
      const selectedDashboard = updateSelectedDashboard(
        state.selectedDashboard,
        dashboards
      );
      return {
        ...state,
        error: null,
        availableDashboardsLoaded: true,
        dashboards,
        dashboardsMap,
        selectedDashboard: updateSelectedDashboard(
          state.selectedDashboard,
          dashboards
        ),
        dashboard:
          selectedDashboard &&
          state.dashboard &&
          state.dashboard.name === selectedDashboard.full_name
            ? state.dashboard
            : null,
      };
    case DashboardActions.EXECUTION_STARTED: {
      const originalDashboard = action.dashboard_node;
      let dashboard;
      // For benchmarks and controls that are run directly from a mod,
      // we need to wrap these in an artificial dashboard, so we can treat
      // it just like any other dashboard
      if (action.dashboard_node.node_type !== "dashboard") {
        dashboard = wrapDefinitionInArtificialDashboard(originalDashboard);
      } else {
        dashboard = addDataToDashboard(action.dashboard_node, state.sqlDataMap);
      }
      return {
        ...state,
        error: null,
        dashboard,
        execution_id: action.execution_id,
        refetchDashboard: false,
        state: "running",
      };
    }
    case DashboardActions.EXECUTION_COMPLETE: {
      // If we're in live mode and not expecting execution events for this ID
      if (
        state.dataMode === "live" &&
        action.execution_id !== state.execution_id
      ) {
        return state;
      }

      const originalDashboard = action.dashboard_node;
      let dashboard;

      if (action.dashboard_node.type !== "dashboard") {
        dashboard = wrapDefinitionInArtificialDashboard(originalDashboard);
      } else {
        dashboard = originalDashboard;
      }

      // Build map of SQL to data
      const sqlDataMap = buildSqlDataMap(action.dashboard_node);
      // Replace the whole dashboard as this event contains everything
      return {
        ...state,
        error: null,
        dashboard,
        sqlDataMap,
        state: "complete",
      };
    }
    case DashboardActions.EXECUTION_ERROR:
      return { ...state, error: action.error };
    case DashboardActions.CONTROL_COMPLETE:
    case DashboardActions.CONTROL_ERROR:
      // We're not expecting execution events for this ID
      if (action.execution_id !== state.execution_id) {
        return state;
      }

      const updatedDashboard = updateDashboardWithControlEvent(
        state.dashboard,
        action
      );

      if (!updatedDashboard) {
        return state;
      }

      return {
        ...state,
        dashboard: updatedDashboard,
      };
    case DashboardActions.LEAF_NODE_COMPLETE: {
      // We're not expecting execution events for this ID
      if (action.execution_id !== state.execution_id) {
        return state;
      }
      // Find the path to the name key that matches this panel and replace it
      const { dashboard_node } = action;
      let panelPath: string = findPathDeep(
        state.dashboard,
        (v, k) => k === "name" && v === dashboard_node.name
      );

      if (!panelPath) {
        console.warn(
          "Cannot find dashboard panel to update",
          dashboard_node.name
        );
        return state;
      }

      panelPath = panelPath.replace(".name", "");
      let newDashboard = {
        ...state.dashboard,
      };
      newDashboard = set(newDashboard, panelPath, dashboard_node);

      return {
        ...state,
        dashboard: newDashboard,
      };
    }
    case DashboardActions.SELECT_PANEL:
      return { ...state, selectedPanel: action.panel };
    case DashboardActions.CLEAR_SNAPSHOT:
      return { ...state, selectedSnapshot: null, dataMode: "live" };
    case DashboardActions.SELECT_SNAPSHOT:
      return {
        ...state,
        selectedSnapshot: action.snapshot,
        dataMode: "snapshot",
      };
    case DashboardActions.SET_DATA_MODE:
      return {
        ...state,
        dataMode: action.mode,
      };
    case DashboardActions.SET_REFETCH_DASHBOARD:
      return {
        ...state,
        refetchDashboard: true,
      };
    case DashboardActions.SET_DASHBOARD:
      return {
        ...state,
        dashboard: action.dashboard,
      };
    case DashboardActions.SELECT_DASHBOARD:
      return {
        ...state,
        dataMode: "live",
        dashboard: null,
        execution_id: null,
        state: null,
        selectedDashboard: action.dashboard,
        selectedPanel: null,
        selectedSnapshot: null,
        lastChangedInput: null,
      };
    case DashboardActions.CLEAR_DASHBOARD_INPUTS:
      return {
        ...state,
        selectedDashboardInputs: {},
        lastChangedInput: null,
        recordInputsHistory: !!action.recordInputsHistory,
      };
    case DashboardActions.DELETE_DASHBOARD_INPUT:
      const { [action.name]: toDelete, ...rest } =
        state.selectedDashboardInputs;
      return {
        ...state,
        selectedDashboardInputs: {
          ...rest,
        },
        lastChangedInput: action.name,
        recordInputsHistory: !!action.recordInputsHistory,
      };
    case DashboardActions.SET_DASHBOARD_INPUT:
      return {
        ...state,
        selectedDashboardInputs: {
          ...state.selectedDashboardInputs,
          [action.name]: action.value,
        },
        lastChangedInput: action.name,
        recordInputsHistory: !!action.recordInputsHistory,
      };
    case DashboardActions.SET_DASHBOARD_INPUTS:
      return {
        ...state,
        selectedDashboardInputs: action.value,
        lastChangedInput: null,
        recordInputsHistory: !!action.recordInputsHistory,
      };
    case DashboardActions.INPUT_VALUES_CLEARED: {
      // We're not expecting execution events for this ID
      if (action.execution_id !== state.execution_id) {
        return state;
      }
      const newSelectedDashboardInputs = { ...state.selectedDashboardInputs };
      for (const input of action.cleared_inputs || []) {
        delete newSelectedDashboardInputs[input];
      }
      return {
        ...state,
        selectedDashboardInputs: newSelectedDashboardInputs,
        lastChangedInput: null,
        recordInputsHistory: false,
      };
    }
    case DashboardActions.SET_DASHBOARD_SEARCH_VALUE:
      return {
        ...state,
        search: {
          ...state.search,
          value: action.value,
        },
      };
    case DashboardActions.SET_DASHBOARD_SEARCH_GROUP_BY:
      return {
        ...state,
        search: {
          ...state.search,
          groupBy: {
            value: action.value,
            tag: action.tag,
          },
        },
      };
    case DashboardActions.SET_DASHBOARD_TAG_KEYS:
      return {
        ...state,
        dashboardTags: {
          ...state.dashboardTags,
          keys: action.keys,
        },
      };
    case DashboardActions.WORKSPACE_ERROR:
      return { ...state, error: action.error };
    default:
      console.warn(`Unsupported action ${action.type}`, action);
      return state;
  }
}

const buildSelectedDashboardInputsFromSearchParams = (searchParams) => {
  const selectedDashboardInputs = {};
  // @ts-ignore
  for (const entry of searchParams.entries()) {
    if (!entry[0].startsWith("input")) {
      continue;
    }
    selectedDashboardInputs[entry[0]] = entry[1];
  }
  return selectedDashboardInputs;
};

const getInitialState = (searchParams) => {
  return {
    availableDashboardsLoaded: false,
    metadata: null,
    dashboards: [],
    dashboardTags: {
      keys: [],
    },
    dataMode: searchParams.get("mode") || "live",
    refetchDashboard: false,
    error: null,

    dashboard: null,
    selectedPanel: null,
    selectedDashboard: null,
    selectedDashboardInputs:
      buildSelectedDashboardInputsFromSearchParams(searchParams),
    selectedSnapshot: null,
    lastChangedInput: null,

    search: {
      value: searchParams.get("search") || "",
      groupBy: {
        value: searchParams.get("group_by") || "tag",
        tag: searchParams.get("tag") || "service",
      },
    },

    sqlDataMap: {},

    execution_id: null,
  };
};

const DashboardContext = createContext<IDashboardContext | null>(null);

const DashboardProvider = ({
  analyticsContext,
  breakpointContext,
  children,
  componentOverrides = {},
  eventHooks = {},
  socketFactory,
  themeContext,
}) => {
  const components = buildComponentsMap(componentOverrides);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [state, dispatch] = useReducer(reducer, getInitialState(searchParams));
  const { dashboard_name } = useParams();
  const { ready: socketReady, send: sendSocketMessage } = useDashboardWebSocket(
    dispatch,
    socketFactory,
    eventHooks
  );
  const {
    setMetadata: setAnalyticsMetadata,
    setSelectedDashboard: setAnalyticsSelectedDashboard,
  } = analyticsContext;

  const location = useLocation();
  const navigationType = useNavigationType();

  // Keep track of the previous selected dashboard and inputs
  const previousSelectedDashboardStates: SelectedDashboardStates | undefined =
    usePrevious({
      searchParams,
      dashboard_name,
      dataMode: state.dataMode,
      search: state.search,
      selectedDashboard: state.selectedDashboard,
      selectedDashboardInputs: state.selectedDashboardInputs,
      selectedSnapshot: state.selectedSnapshot,
    });

  // console.log({
  //   selectedDashboard: state.selectedDashboard,
  //   selectedSnapshot: state.selectedSnapshot,
  //   dashboard: state.dashboard,
  //   dataMode: state.dataMode,
  //   refetchDashboard: state.refetchDashboard,
  // });

  // Initial sync into URL
  useEffect(() => {
    if (searchParams.has("mode")) {
      return;
    }
    searchParams.set("mode", state.dataMode);
    setSearchParams(searchParams, { replace: true });
  }, []);

  // Alert analytics
  useEffect(() => {
    setAnalyticsMetadata(state.metadata);
  }, [state.metadata, setAnalyticsMetadata]);

  useEffect(() => {
    setAnalyticsSelectedDashboard(state.selectedDashboard);
  }, [state.selectedDashboard, setAnalyticsSelectedDashboard]);

  // Ensure that on history pop / push we sync the new values into state
  useEffect(() => {
    if (navigationType !== "POP" && navigationType !== "PUSH") {
      return;
    }
    if (location.key === "default") {
      return;
    }

    // If we've just popped or pushed from one dashboard to another, then we don't want to add the search to the URL
    // as that will show the dashboard list, but we want to see the dashboard that we came from / went to previously.
    const goneFromDashboardToDashboard =
      // @ts-ignore
      previousSelectedDashboardStates?.dashboard_name &&
      dashboard_name &&
      // @ts-ignore
      previousSelectedDashboardStates.dashboard_name !== dashboard_name;

    const search = searchParams.get("search") || "";
    const groupBy = searchParams.get("group_by") || "tag";
    const tag = searchParams.get("tag") || "service";
    const dataMode = searchParams.get("mode") || "live";
    const inputs = buildSelectedDashboardInputsFromSearchParams(searchParams);
    dispatch({
      type: DashboardActions.SET_DASHBOARD_SEARCH_VALUE,
      value: goneFromDashboardToDashboard ? "" : search,
    });
    dispatch({
      type: DashboardActions.SET_DASHBOARD_SEARCH_GROUP_BY,
      value: groupBy,
      tag,
    });
    dispatch({
      type: DashboardActions.SET_DASHBOARD_INPUTS,
      value: inputs,
      recordInputsHistory: false,
    });
    dispatch({
      type: DashboardActions.SET_DATA_MODE,
      dataMode,
    });
  }, [
    dashboard_name,
    dispatch,
    location,
    navigationType,
    previousSelectedDashboardStates,
    searchParams,
  ]);

  useEffect(() => {
    // If no search params have changed
    if (
      previousSelectedDashboardStates &&
      // @ts-ignore
      previousSelectedDashboardStates?.dashboard_name === dashboard_name &&
      // @ts-ignore
      previousSelectedDashboardStates.dataMode === state.dataMode &&
      // @ts-ignore
      previousSelectedDashboardStates.search.value === state.search.value &&
      // @ts-ignore
      previousSelectedDashboardStates.search.groupBy.value ===
        state.search.groupBy.value &&
      // @ts-ignore
      previousSelectedDashboardStates.search.groupBy.tag ===
        state.search.groupBy.tag &&
      // @ts-ignore
      previousSelectedDashboardStates.searchParams.toString() ===
        searchParams.toString()
    ) {
      return;
    }

    const {
      value: searchValue,
      groupBy: { value: groupByValue, tag },
    } = state.search;

    if (dashboard_name) {
      // Only set group_by and tag if we have a search
      if (searchValue) {
        searchParams.set("search", searchValue);
        searchParams.set("group_by", groupByValue);

        if (groupByValue === "mod") {
          searchParams.delete("tag");
        } else if (groupByValue === "tag") {
          searchParams.set("tag", tag);
        } else {
          searchParams.delete("group_by");
          searchParams.delete("tag");
        }
      } else {
        searchParams.delete("search");
        searchParams.delete("group_by");
        searchParams.delete("tag");
      }
    } else {
      if (searchValue) {
        searchParams.set("search", searchValue);
      } else {
        searchParams.delete("search");
      }

      searchParams.set("group_by", groupByValue);

      if (groupByValue === "mod") {
        searchParams.delete("tag");
      } else if (groupByValue === "tag") {
        searchParams.set("tag", tag);
      } else {
        searchParams.delete("group_by");
        searchParams.delete("tag");
      }
    }

    searchParams.set("mode", state.dataMode);
    setSearchParams(searchParams, { replace: true });
  }, [
    previousSelectedDashboardStates,
    dashboard_name,
    searchParams,
    setSearchParams,
    state.dataMode,
    state.search,
  ]);

  useEffect(() => {
    // If we've got no dashboard selected in the URL, but we've got one selected in state,
    // then clear both the inputs and the selected dashboard in state
    if (!dashboard_name && state.selectedDashboard) {
      dispatch({
        type: DashboardActions.CLEAR_DASHBOARD_INPUTS,
        recordInputsHistory: false,
      });
      dispatch({
        type: DashboardActions.SELECT_DASHBOARD,
        dashboard: null,
        recordInputsHistory: false,
      });
      return;
    }
    // Else if we've got a dashboard selected in the URL and don't have one selected in state,
    // select that dashboard
    if (dashboard_name && !state.selectedDashboard) {
      const dashboard = state.dashboards.find(
        (dashboard) => dashboard.full_name === dashboard_name
      );
      dispatch({ type: DashboardActions.SELECT_DASHBOARD, dashboard });
      return;
    }
    // Else if we've changed to a different report in the URL then clear the inputs and select the
    // dashboard in state
    if (
      dashboard_name &&
      state.selectedDashboard &&
      dashboard_name !== state.selectedDashboard.full_name
    ) {
      const dashboard = state.dashboards.find(
        (dashboard) => dashboard.full_name === dashboard_name
      );
      dispatch({ type: DashboardActions.SELECT_DASHBOARD, dashboard });
      const value = buildSelectedDashboardInputsFromSearchParams(searchParams);
      dispatch({
        type: DashboardActions.SET_DASHBOARD_INPUTS,
        value,
        recordInputsHistory: false,
      });
    }
  }, [dashboard_name, searchParams, state.dashboards, state.selectedDashboard]);

  useEffect(() => {
    // This effect will send events over websockets and depends on there being a dashboard selected
    if (!socketReady || !state.selectedDashboard) {
      return;
    }

    // If we didn't previously have a dashboard selected in state (e.g. you've gone from home page
    // to a report, or it's first load), or the selected dashboard has been changed, select that
    // report over the socket
    if (
      state.dataMode === "live" &&
      (!previousSelectedDashboardStates ||
        // @ts-ignore
        !previousSelectedDashboardStates.selectedDashboard ||
        state.selectedDashboard.full_name !==
          // @ts-ignore
          previousSelectedDashboardStates.selectedDashboard.full_name ||
        // @ts-ignore
        (!previousSelectedDashboardStates.refetchDashboard &&
          state.refetchDashboard))
    ) {
      sendSocketMessage({
        action: SocketActions.CLEAR_DASHBOARD,
      });
      sendSocketMessage({
        action: SocketActions.SELECT_DASHBOARD,
        payload: {
          dashboard: {
            full_name: state.selectedDashboard.full_name,
          },
          input_values: state.selectedDashboardInputs,
        },
      });
      return;
    }
    // Else if we did previously have a dashboard selected in state and the
    // inputs have changed, then update the inputs over the socket
    if (
      state.dataMode === "live" &&
      previousSelectedDashboardStates &&
      // @ts-ignore
      previousSelectedDashboardStates.selectedDashboard &&
      !isEqual(
        // @ts-ignore
        previousSelectedDashboardStates.selectedDashboardInputs,
        state.selectedDashboardInputs
      )
    ) {
      sendSocketMessage({
        action: SocketActions.INPUT_CHANGED,
        payload: {
          dashboard: {
            full_name: state.selectedDashboard.full_name,
          },
          changed_input: state.lastChangedInput,
          input_values: state.selectedDashboardInputs,
        },
      });
    }
  }, [
    previousSelectedDashboardStates,
    sendSocketMessage,
    socketReady,
    state.selectedDashboard,
    state.selectedDashboardInputs,
    state.lastChangedInput,
    state.dataMode,
    state.refetchDashboard,
  ]);

  useEffect(() => {
    // This effect will send events over websockets and depends on there being no dashboard selected
    if (!socketReady || state.selectedDashboard) {
      return;
    }

    // If we've gone from having a report selected, to having nothing selected, clear the dashboard state
    if (
      previousSelectedDashboardStates &&
      // @ts-ignore
      previousSelectedDashboardStates.selectedDashboard
    ) {
      sendSocketMessage({
        action: SocketActions.CLEAR_DASHBOARD,
      });
    }
  }, [
    previousSelectedDashboardStates,
    sendSocketMessage,
    socketReady,
    state.selectedDashboard,
  ]);

  useEffect(() => {
    // Don't do anything as this is handled elsewhere
    if (navigationType === "POP" || navigationType === "PUSH") {
      return;
    }

    if (!previousSelectedDashboardStates) {
      return;
    }

    if (
      isEqual(
        state.selectedDashboardInputs,
        // @ts-ignore
        previousSelectedDashboardStates.selectedDashboardInputs
      )
    ) {
      return;
    }

    // Only record history when it's the same report before and after and the inputs have changed
    const shouldRecordHistory =
      state.recordInputsHistory &&
      // @ts-ignore
      !!previousSelectedDashboardStates.selectedDashboard &&
      !!state.selectedDashboard &&
      // @ts-ignore
      previousSelectedDashboardStates.selectedDashboard.full_name ===
        state.selectedDashboard.full_name;

    // Sync params into the URL
    setSearchParams(
      {
        ...state.selectedDashboardInputs,
        mode: state.dataMode,
      },
      {
        replace: !shouldRecordHistory,
      }
    );
  }, [
    navigationType,
    previousSelectedDashboardStates,
    setSearchParams,
    state.dataMode,
    state.recordInputsHistory,
    state.selectedDashboard,
    state.selectedDashboardInputs,
  ]);

  useEffect(() => {
    if (!state.availableDashboardsLoaded || !dashboard_name) {
      return;
    }
    // If the dashboard we're viewing no longer exists, go back to the main page
    if (!state.dashboards.find((r) => r.full_name === dashboard_name)) {
      navigate("../", { replace: true });
    }
  }, [
    navigate,
    dashboard_name,
    state.availableDashboardsLoaded,
    state.dashboards,
  ]);

  useEffect(() => {
    if (!state.selectedDashboard) {
      document.title = "Dashboards | Steampipe";
    } else {
      document.title = `${
        state.selectedDashboard.title || state.selectedDashboard.full_name
      } | Dashboards | Steampipe`;
    }
  }, [state.selectedDashboard]);

  const [hotKeysHandlers, setHotKeysHandlers] = useState({
    CLOSE_PANEL_DETAIL: noop,
  });

  const hotKeysMap = {
    CLOSE_PANEL_DETAIL: ["esc"],
  };

  const closePanelDetail = useCallback(() => {
    dispatch({
      type: DashboardActions.SELECT_PANEL,
      panel: null,
    });
  }, []);

  useEffect(() => {
    setHotKeysHandlers({
      CLOSE_PANEL_DETAIL: closePanelDetail,
    });
  }, [closePanelDetail]);

  return (
    <DashboardContext.Provider
      value={{
        ...state,
        analyticsContext,
        breakpointContext,
        components,
        dispatch,
        closePanelDetail,
        themeContext,
      }}
    >
      <GlobalHotKeys
        allowChanges
        keyMap={hotKeysMap}
        handlers={hotKeysHandlers}
      />
      {children}
    </DashboardContext.Provider>
  );
};

const useDashboard = () => {
  const context = useContext(DashboardContext);
  if (context === undefined) {
    throw new Error("useDashboard must be used within a DashboardContext");
  }
  return context as IDashboardContext;
};

export { DashboardActions, DashboardContext, DashboardProvider, useDashboard };
