import get from "lodash/get";
import isEqual from "lodash/isEqual";
import paths from "deepdash/paths";
import set from "lodash/set";
import sortBy from "lodash/sortBy";
import useDashboardWebSocket, { SocketActions } from "./useDashboardWebSocket";
import usePrevious from "./usePrevious";
import { buildComponentsMap } from "../components";
import {
  createContext,
  Ref,
  useCallback,
  useContext,
  useEffect,
  useReducer,
  useState,
} from "react";
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
  wrapperRef: Ref<null>;
}

export interface ComponentsMap {
  [name: string]: any;
}

export interface PanelsMap {
  [name: string]: PanelDefinition;
}

export type DashboardDataMode = "live" | "snapshot";

export type DashboardRunState = "running" | "complete";

interface IDashboardContext {
  metadata: DashboardMetadata | null;
  availableDashboardsLoaded: boolean;

  closePanelDetail(): void;
  dispatch(action: DashboardAction): void;

  dataMode: DashboardDataMode;
  snapshotId: string | null;

  refetchDashboard: boolean;

  error: any;

  panelsMap: PanelsMap;

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

  state: DashboardRunState;
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
  error?: Error;
  properties?: PanelProperties;
  dashboard: string;
}

export interface BenchmarkDefinition extends PanelDefinition {
  children?: BenchmarkDefinition | ControlDefinition[];
  description?: string;
}

export interface ControlDefinition extends PanelDefinition {}

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

interface DashboardProviderProps {
  analyticsContext: any;
  breakpointContext: any;
  children: null | JSX.Element | JSX.Element[];
  componentOverrides?: {};
  eventHooks?: {};
  featureFlags?: string[];
  socketFactory?: () => WebSocket;
  stateDefaults?: {};
  themeContext: any;
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

function buildSqlDataMap(panels: PanelsMap): SQLDataMap {
  const sqlPaths = paths(panels, { leavesOnly: true }).filter((path) =>
    path.endsWith(".sql")
  );
  const sqlDataMap = {};
  for (const sqlPath of sqlPaths) {
    // @ts-ignore
    const sql: string = get(panels, sqlPath);
    const dataPath = `${sqlPath.substring(0, sqlPath.indexOf(".sql"))}.data`;
    const data = get(panels, dataPath);
    if (!sqlDataMap[sql]) {
      sqlDataMap[sql] = data;
    }
  }
  return sqlDataMap;
}

function addDataToPanels(panels: PanelsMap, sqlDataMap: SQLDataMap): PanelsMap {
  const sqlPaths = paths(panels, { leavesOnly: true }).filter((path) =>
    path.endsWith(".sql")
  );
  for (const sqlPath of sqlPaths) {
    // @ts-ignore
    const sql: string = get(panels, sqlPath);
    const data = sqlDataMap[sql];
    if (!data) {
      continue;
    }
    const dataPath = `${sqlPath.substring(0, sqlPath.indexOf(".sql"))}.data`;
    set(panels, dataPath, data);
  }
  return panels;
}

const wrapDefinitionInArtificialDashboard = (
  definition: DashboardDefinition,
  layout: any
): DashboardDefinition => {
  const { title: defTitle, ...definitionWithoutTitle } = definition;
  const { title: layoutTitle, ...layoutWithoutTitle } = layout;
  return {
    artificial: true,
    name: definition.name,
    title: definition.title,
    node_type: "dashboard",
    children: [
      {
        ...definitionWithoutTitle,
        ...layoutWithoutTitle,
      },
    ],
    dashboard: definition.dashboard,
  };
};

const updatePanelsMapWithControlEvent = (panelsMap, action) => {
  return {
    ...panelsMap,
    [action.control.name]: action.control,
  };
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
        dashboard = wrapDefinitionInArtificialDashboard(
          originalDashboard,
          action.layout
        );
      } else {
        dashboard = {
          ...originalDashboard,
          ...action.layout,
        };
      }

      return {
        ...state,
        error: null,
        panelsMap: addDataToPanels(
          action.panels || action.leaf_nodes,
          state.sqlDataMap
        ),
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

      if (action.dashboard_node.node_type !== "dashboard") {
        dashboard = wrapDefinitionInArtificialDashboard(
          originalDashboard,
          action.layout
        );
      } else {
        dashboard = {
          ...originalDashboard,
          ...action.layout,
        };
      }

      // Build map of SQL to data
      const sqlDataMap = buildSqlDataMap(action.panels);
      // Replace the whole dashboard as this event contains everything
      return {
        ...state,
        error: null,
        panelsMap: action.panels,
        dashboard,
        sqlDataMap,
        state: "complete",
      };
    }
    case DashboardActions.EXECUTION_ERROR:
      return { ...state, error: action.error, state: "error" };
    case DashboardActions.CONTROL_COMPLETE:
    case DashboardActions.CONTROL_ERROR:
      // We're not expecting execution events for this ID
      if (action.execution_id !== state.execution_id) {
        return state;
      }

      const updatedPanelsMap = updatePanelsMapWithControlEvent(
        state.panelsMap,
        action
      );

      if (!updatedPanelsMap) {
        return state;
      }

      return {
        ...state,
        panelsMap: updatedPanelsMap,
      };
    case DashboardActions.LEAF_NODE_COMPLETE: {
      // We're not expecting execution events for this ID
      if (action.execution_id !== state.execution_id) {
        return state;
      }

      const { dashboard_node } = action;

      const panelsMap = {
        ...state.panelsMap,
        [dashboard_node.name]: dashboard_node,
      };

      return {
        ...state,
        panelsMap,
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
        dataMode: action.dataMode,
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
        dataMode: action.dataMode || "live",
        dashboard: null,
        execution_id: null,
        snapshotId: null,
        state: null,
        selectedDashboard: action.dashboard,
        selectedPanel: null,
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

const getInitialState = (searchParams, defaults = {}) => {
  return {
    availableDashboardsLoaded: false,
    metadata: null,
    dashboards: [],
    dashboardTags: {
      keys: [],
    },
    dataMode: searchParams.get("mode") || "live",
    snapshotId: searchParams.has("snapshot_id")
      ? searchParams.get("snapshot_id")
      : null,
    refetchDashboard: false,
    error: null,

    panelsMap: {},
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
        value:
          searchParams.get("group_by") ||
          get(defaults, "search.groupBy.value", "tag"),
        tag:
          searchParams.get("tag") ||
          get(defaults, "search.groupBy.value", "service"),
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
  featureFlags = [],
  socketFactory,
  stateDefaults = {},
  themeContext,
}: DashboardProviderProps) => {
  const components = buildComponentsMap(componentOverrides);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [state, dispatchInner] = useReducer(
    reducer,
    getInitialState(searchParams, stateDefaults)
  );
  const dispatch = useCallback((action) => {
    console.log(action.type, action);
    dispatchInner(action);
  }, []);
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
      refetchDashboard: state.refetchDashboard,
      search: state.search,
      selectedDashboard: state.selectedDashboard,
      selectedDashboardInputs: state.selectedDashboardInputs,
      selectedSnapshot: state.selectedSnapshot,
    });

  // Initial sync into URL
  useEffect(() => {
    if (
      !featureFlags.includes("snapshots") ||
      (searchParams.has("mode") && searchParams.get("mode") === state.dataMode)
    ) {
      return;
    }
    searchParams.set("mode", state.dataMode);
    setSearchParams(searchParams, { replace: true });
  }, [featureFlags, searchParams, setSearchParams, state.dataMode]);

  useEffect(() => {
    if (featureFlags.includes("snapshots") && state.selectedSnapshot) {
      searchParams.set("snapshot_id", state.selectedSnapshot.id);
      setSearchParams(searchParams, { replace: true });
    }
  }, [featureFlags, searchParams, setSearchParams, state.selectedSnapshot]);

  useEffect(() => {
    if (
      featureFlags.includes("snapshots") &&
      state.dataMode === "live" &&
      searchParams.has("snapshot_id")
    ) {
      searchParams.delete("snapshot_id");
      setSearchParams(searchParams, { replace: true });
    }
  }, [featureFlags, searchParams, setSearchParams, state.dataMode]);

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
    const groupBy =
      searchParams.get("group_by") ||
      get(stateDefaults, "search.groupBy.value", "tag");
    const tag =
      searchParams.get("tag") ||
      get(stateDefaults, "search.groupBy.tag", "service");
    const dataMode = searchParams.has("mode")
      ? searchParams.get("mode")
      : "live";
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
    if (featureFlags.includes("snapshots")) {
      dispatch({
        type: DashboardActions.SET_DATA_MODE,
        dataMode,
      });
    }
  }, [
    dashboard_name,
    dispatch,
    featureFlags,
    location,
    navigationType,
    previousSelectedDashboardStates,
    searchParams,
    stateDefaults,
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

    if (featureFlags.includes("snapshots")) {
      searchParams.set("mode", state.dataMode);
    }
    setSearchParams(searchParams, { replace: true });
  }, [
    dashboard_name,
    featureFlags,
    previousSelectedDashboardStates,
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
      dispatch({
        type: DashboardActions.SELECT_DASHBOARD,
        dashboard,
        dataMode: state.dataMode,
      });
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
  }, [
    dashboard_name,
    dispatch,
    searchParams,
    state.dashboards,
    state.dataMode,
    state.selectedDashboard,
  ]);

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
    const newParams = {
      ...state.selectedDashboardInputs,
    };
    if (featureFlags.includes("snapshots")) {
      newParams.mode = state.dataMode;
    }
    setSearchParams(newParams, {
      replace: !shouldRecordHistory,
    });
  }, [
    featureFlags,
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
  }, [dispatch]);

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
