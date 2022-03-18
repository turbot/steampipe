import findPathDeep from "deepdash/findPathDeep";
import paths from "deepdash/paths";
import useDashboardWebSocket, { SocketActions } from "./useDashboardWebSocket";
import usePrevious from "./usePrevious";
import { CheckLeafNodeExecutionTree } from "../components/dashboards/check/common";
import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useReducer,
  useState,
} from "react";
import { Theme } from "./useTheme";
import { get, isEqual, set, sortBy } from "lodash";
import { GlobalHotKeys } from "react-hotkeys";
import { LeafNodeData } from "../components/dashboards/common";
import { noop } from "../utils/func";
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

interface IDashboardContext {
  metadata: DashboardMetadata | null;
  availableDashboardsLoaded: boolean;

  closePanelDetail(): void;
  dispatch(DispatchAction): void;

  error: any;

  dashboards: AvailableDashboard[];
  dashboard: DashboardDefinition | null;

  selectedPanel: PanelDefinition | null;
  selectedDashboard: AvailableDashboard | null;
  selectedDashboardInputs: DashboardInputs;
  lastChangedInput: string | null;

  sqlDataMap: SQLDataMap;

  dashboardTags: DashboardTags;

  search: DashboardSearch;

  breakpointContext: IBreakpointContext;
  themeContext: IThemeContext;
}

export interface IActions {
  [type: string]: string;
}

const DashboardActions: IActions = {
  AVAILABLE_DASHBOARDS: "available_dashboards",
  CLEAR_DASHBOARD_INPUTS: "clear_dashboard_inputs",
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
  SET_DASHBOARD_INPUT: "set_dashboard_input",
  SET_DASHBOARD_INPUTS: "set_dashboard_inputs",
  SET_DASHBOARD_SEARCH_VALUE: "set_dashboard_search_value",
  SET_DASHBOARD_SEARCH_GROUP_BY: "set_dashboard_search_group_by",
  SET_DASHBOARD_TAG_KEYS: "set_dashboard_tag_keys",
  WORKSPACE_ERROR: "workspace_error",
};

const dashboardActions = Object.values(DashboardActions);

// https://github.com/microsoft/TypeScript/issues/28046
export type ElementType<T extends ReadonlyArray<unknown>> =
  T extends ReadonlyArray<infer ElementType> ? ElementType : never;
type DashboardActionType = ElementType<typeof dashboardActions>;

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
  dashboardName: string | null;
  search: DashboardSearch;
  selectedDashboard: AvailableDashboard | null;
  selectedDashboardInputs: DashboardInputs;
}

interface DashboardInputs {
  [name: string]: string;
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

interface AvailableDashboardTags {
  [key: string]: string;
}

export interface AvailableDashboard {
  full_name: string;
  short_name: string;
  mod_full_name: string;
  tags: AvailableDashboardTags;
  title: string;
}

interface AvailableDashboardsForModDictionary {
  [key: string]: AvailableDashboard;
}

interface DashboardsByModDictionary {
  [key: string]: AvailableDashboardsForModDictionary;
}

export interface ContainerDefinition {
  name: string;
  node_type?: string;
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
  width?: number;
  sql?: string;
  data?: LeafNodeData;
  source_definition?: string;
  execution_tree?: CheckLeafNodeExecutionTree;
  error?: Error;
  properties?: PanelProperties;
  dashboard: string;
}

export interface DashboardDefinition {
  name: string;
  title?: string;
  width?: number;
  children?: (ContainerDefinition | PanelDefinition)[];
  dashboard: string;
}

const buildDashboardsList = (
  dashboards_by_mod: DashboardsByModDictionary
): AvailableDashboard[] => {
  const dashboards: AvailableDashboard[] = [];
  for (const [mod_full_name, dashboards_for_mod] of Object.entries(
    dashboards_by_mod
  )) {
    for (const [, dashboard] of Object.entries(dashboards_for_mod)) {
      dashboards.push({
        title: dashboard.title,
        full_name: dashboard.full_name,
        short_name: dashboard.short_name,
        tags: dashboard.tags,
        mod_full_name: mod_full_name,
      });
    }
  }
  return sortBy(dashboards, [
    (dashboard) =>
      dashboard.title
        ? dashboard.title.toLowerCase()
        : dashboard.full_name.toLowerCase(),
  ]);
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

function reducer(state, action) {
  switch (action.type) {
    case DashboardActions.DASHBOARD_METADATA:
      return {
        ...state,
        metadata: action.metadata,
      };
    case DashboardActions.AVAILABLE_DASHBOARDS:
      const dashboards = buildDashboardsList(action.dashboards_by_mod);
      const selectedDashboard = updateSelectedDashboard(
        state.selectedDashboard,
        dashboards
      );
      return {
        ...state,
        error: null,
        availableDashboardsLoaded: true,
        dashboards,
        selectedDashboard: updateSelectedDashboard(
          state.selectedDashboard,
          dashboards
        ),
        dashboard:
          selectedDashboard &&
          state.dashboard.name === selectedDashboard.full_name
            ? state.dashboard
            : null,
      };
    case DashboardActions.EXECUTION_STARTED:
      const dashboardWithData = addDataToDashboard(
        action.dashboard_node,
        state.sqlDataMap
      );
      return {
        ...state,
        error: null,
        dashboard: dashboardWithData,
        state: "running",
      };
    case DashboardActions.EXECUTION_COMPLETE:
      // Build map of SQL to data
      const sqlDataMap = buildSqlDataMap(action.dashboard_node);
      // Replace the whole dashboard as this event contains everything
      return {
        ...state,
        error: null,
        dashboard: action.dashboard_node,
        sqlDataMap,
        state: "complete",
      };
    case DashboardActions.EXECUTION_ERROR:
      // console.error("Got execution error", action);
      return state;
    case DashboardActions.LEAF_NODE_PROGRESS:
    case DashboardActions.LEAF_NODE_COMPLETE: {
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
    case DashboardActions.SELECT_DASHBOARD:
      return {
        ...state,
        dashboard: null,
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
    error: null,

    dashboard: null,
    selectedPanel: null,
    selectedDashboard: null,
    selectedDashboardInputs:
      buildSelectedDashboardInputsFromSearchParams(searchParams),
    lastChangedInput: null,

    search: {
      value: searchParams.get("search") || "",
      groupBy: {
        value: searchParams.get("group_by") || "tag",
        tag: searchParams.get("tag") || "service",
      },
    },

    sqlDataMap: {},
  };
};

const DashboardContext = createContext<IDashboardContext | null>(null);

const DashboardProvider = ({
  analyticsContext,
  breakpointContext,
  children,
  socketFactory,
  themeContext,
}) => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [state, dispatch] = useReducer(reducer, getInitialState(searchParams));
  const { dashboardName } = useParams();
  console.log(dashboardName);
  const { ready: socketReady, send: sendSocketMessage } = useDashboardWebSocket(
    dispatch,
    socketFactory
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
      dashboardName,
      search: state.search,
      selectedDashboard: state.selectedDashboard,
      selectedDashboardInputs: state.selectedDashboardInputs,
    });

  // Alert analytics
  useEffect(() => {
    setAnalyticsMetadata(state.metadata);
  }, [state.metadata, setAnalyticsMetadata]);
  useEffect(() => {
    setAnalyticsSelectedDashboard(state.metadata);
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
      previousSelectedDashboardStates?.dashboardName &&
      dashboardName &&
      // @ts-ignore
      previousSelectedDashboardStates.dashboardName !== dashboardName;

    const search = searchParams.get("search") || "";
    const groupBy = searchParams.get("group_by") || "tag";
    const tag = searchParams.get("tag") || "service";
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
  }, [
    dashboardName,
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
      previousSelectedDashboardStates?.dashboardName === dashboardName &&
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

    if (dashboardName) {
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

    setSearchParams(searchParams, { replace: true });
  }, [
    previousSelectedDashboardStates,
    dashboardName,
    searchParams,
    setSearchParams,
    state.search,
  ]);

  useEffect(() => {
    // If we've got no dashboard selected in the URL, but we've got one selected in state,
    // then clear both the inputs and the selected dashboard in state
    if (!dashboardName && state.selectedDashboard) {
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
    if (dashboardName && !state.selectedDashboard) {
      const dashboard = state.dashboards.find(
        (dashboard) => dashboard.full_name === dashboardName
      );
      dispatch({ type: DashboardActions.SELECT_DASHBOARD, dashboard });
      return;
    }
    // Else if we've changed to a different report in the URL then clear the inputs and select the
    // dashboard in state
    if (
      dashboardName &&
      state.selectedDashboard &&
      dashboardName !== state.selectedDashboard.full_name
    ) {
      const dashboard = state.dashboards.find(
        (dashboard) => dashboard.full_name === dashboardName
      );
      dispatch({ type: DashboardActions.SELECT_DASHBOARD, dashboard });
      const value = buildSelectedDashboardInputsFromSearchParams(searchParams);
      dispatch({
        type: DashboardActions.SET_DASHBOARD_INPUTS,
        value,
        recordInputsHistory: false,
      });
    }
  }, [dashboardName, searchParams, state.dashboards, state.selectedDashboard]);

  useEffect(() => {
    // This effect will send events over websockets and depends on there being a dashboard selected
    if (!socketReady || !state.selectedDashboard) {
      return;
    }

    // If we didn't previously have a dashboard selected in state (e.g. you've gone from home page
    // to a report, or it's first load), or the selected dashboard has been changed, select that
    // report over the socket
    if (
      !previousSelectedDashboardStates ||
      // @ts-ignore
      !previousSelectedDashboardStates.selectedDashboard ||
      state.selectedDashboard.full_name !==
        // @ts-ignore
        previousSelectedDashboardStates.selectedDashboard.full_name
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
    setSearchParams(state.selectedDashboardInputs, {
      replace: !shouldRecordHistory,
    });
  }, [
    navigationType,
    previousSelectedDashboardStates,
    setSearchParams,
    state.recordInputsHistory,
    state.selectedDashboard,
    state.selectedDashboardInputs,
  ]);

  useEffect(() => {
    if (!state.availableDashboardsLoaded || !dashboardName) {
      return;
    }
    // If the dashboard we're viewing no longer exists, go back to the main page
    if (!state.dashboards.find((r) => r.full_name === dashboardName)) {
      navigate("../", { replace: true });
    }
  }, [
    navigate,
    dashboardName,
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
