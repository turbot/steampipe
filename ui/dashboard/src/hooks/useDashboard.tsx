import findPathDeep from "deepdash/findPathDeep";
import paths from "deepdash/paths";
import usePrevious from "./usePrevious";
import useDashboardWebSocket, { SocketActions } from "./useDashboardWebSocket";
import { CheckLeafNodeExecutionTree } from "../components/dashboards/check/common";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useReducer,
  useState,
} from "react";
import { FullHeightThemeWrapper } from "./useTheme";
import { get, isEqual, set, sortBy } from "lodash";
import { GlobalHotKeys } from "react-hotkeys";
import { LeafNodeData } from "../components/dashboards/common";
import { noop } from "../utils/func";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";

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
}

export interface IActions {
  [type: string]: string;
}

const DashboardActions: IActions = {
  SET_DASHBOARD_SEARCH_VALUE: "set_dashboard_search_value",
  SET_DASHBOARD_SEARCH_GROUP_BY: "set_dashboard_search_group_by",
  SET_DASHBOARD_TAG_KEYS: "set_dashboard_tag_keys",
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

interface DashboardMetadata {
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
    case "dashboard_metadata":
      return {
        ...state,
        metadata: action.metadata,
      };
    case "available_dashboards":
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
    case "execution_started":
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
    case "execution_complete":
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
    case "leaf_node_progress":
    case "leaf_node_complete": {
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
    case "select_panel":
      return { ...state, selectedPanel: action.panel };
    case "select_dashboard":
      return {
        ...state,
        dashboard: null,
        selectedDashboard: action.dashboard,
        selectedPanel: null,
        lastChangedInput: null,
      };
    case "clear_dashboard_inputs":
      return {
        ...state,
        selectedDashboardInputs: {},
        lastChangedInput: null,
      };
    case "delete_dashboard_input":
      const { [action.name]: toDelete, ...rest } =
        state.selectedDashboardInputs;
      return {
        ...state,
        selectedDashboardInputs: {
          ...rest,
        },
        lastChangedInput: action.name,
      };
    case "set_dashboard_input":
      return {
        ...state,
        selectedDashboardInputs: {
          ...state.selectedDashboardInputs,
          [action.name]: action.value,
        },
        lastChangedInput: action.name,
      };
    case "set_dashboard_inputs":
      return {
        ...state,
        selectedDashboardInputs: action.value,
        lastChangedInput: null,
      };
    case "input_values_cleared": {
      const newSelectedDashboardInputs = { ...state.selectedDashboardInputs };
      for (const input of action.cleared_inputs || []) {
        delete newSelectedDashboardInputs[input];
      }
      return {
        ...state,
        selectedDashboardInputs: newSelectedDashboardInputs,
        lastChangedInput: null,
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
    case "workspace_error":
      return { ...state, error: action.error };
    // Not emitting these from the dashboard server yet
    case "panel_changed":
    case "report_changed":
    case "report_complete":
    case "report_error":
    case "report_event":
      return state;
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

// const initialiseInputs = (
//   initialState: IDashboardContext,
//   searchParams: URLSearchParams
// ) => ({
//   ...initialState,
//   selectedDashboardInputs:
//     buildSelectedDashboardInputsFromSearchParams(searchParams),
// });

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

const DashboardProvider = ({ children }) => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [state, dispatch] = useReducer(
    reducer,
    getInitialState(searchParams)
    // (initialState) => initialiseInputs(initialState, searchParams)
  );
  const { dashboardName } = useParams();
  const { ready: socketReady, send: sendSocketMessage } =
    useDashboardWebSocket(dispatch);

  useEffect(() => {
    // If we've got no dashboard selected in the URL, but we've got one selected in state,
    // then clear both the inputs and the selected dashboard in state
    if (!dashboardName && state.selectedDashboard) {
      dispatch({ type: "clear_dashboard_inputs" });
      dispatch({ type: "select_dashboard", dashboard: null });
    }
    // Else if we've got a dashboard selected in the URL and don't have one selected in state,
    // select that dashboard
    else if (dashboardName && !state.selectedDashboard) {
      const dashboard = state.dashboards.find(
        (dashboard) => dashboard.full_name === dashboardName
      );
      dispatch({ type: "select_dashboard", dashboard });
    }
    // Else if we've changed to a different report in the URL then clear the inputs and select the
    // dashboard in state
    else if (
      dashboardName &&
      state.selectedDashboard &&
      dashboardName !== state.selectedDashboard.full_name
    ) {
      const dashboard = state.dashboards.find(
        (dashboard) => dashboard.full_name === dashboardName
      );
      dispatch({ type: "select_dashboard", dashboard });
      const value = buildSelectedDashboardInputsFromSearchParams(searchParams);
      dispatch({ type: "set_dashboard_inputs", value });
    }
  }, [dashboardName, searchParams, state.dashboards, state.selectedDashboard]);

  // Keep track of the previous selected dashboard and inputs
  const previousSelectedDashboardStates: SelectedDashboardStates | undefined =
    usePrevious({
      selectedDashboard: state.selectedDashboard,
      selectedDashboardInputs: state.selectedDashboardInputs,
    });

  useEffect(() => {
    // This effect will send events over websockets and depends on there being a dashboard selected,
    // so assert that
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
    }
    // Else if we did previously have a dashboard selected in state and the
    // inputs have changed, then update the inputs over the socket
    else if (
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
    state.selectedDashboard,
    state.selectedDashboardInputs,
    state.lastChangedInput,
  ]);

  useEffect(() => {
    // This effect will send events over websockets and depends on there being no dashboard selected,
    // so assert that
    if (!socketReady || state.selectedDashboard) {
      return;
    }

    // If we've gone from having a report selected, to having nothing selected, clear the dashboard state
    if (previousSelectedDashboardStates)
      if (
        previousSelectedDashboardStates &&
        // @ts-ignore
        previousSelectedDashboardStates.selectedDashboard
      ) {
        sendSocketMessage({
          action: SocketActions.CLEAR_DASHBOARD,
        });
      }
  }, [previousSelectedDashboardStates, state.selectedDashboard]);

  /*eslint-disable */
  useEffect(() => {
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

    // Only record history when there's a report before and after and the inputs have changed
    const shouldRecordHistory =
      // @ts-ignore
      !!previousSelectedDashboardStates.selectedDashboard &&
      !!state.selectedDashboard;

    // console.log("Inputs changed", {
    //   previous: {
    //     // @ts-ignore
    //     dashboard: previousSelectedDashboardStates.selectedDashboard,
    //     inputs: JSON.stringify(
    //       // @ts-ignore
    //       previousSelectedDashboardStates.selectedDashboardInputs
    //     ),
    //   },
    //   current: {
    //     dashboard: state.selectedDashboard,
    //     inputs: JSON.stringify(state.selectedDashboardInputs),
    //   },
    //   recordingHistory: shouldRecordHistory,
    // });

    // Sync params into the URL
    setSearchParams(state.selectedDashboardInputs, {
      replace: !shouldRecordHistory,
    });
  }, [
    previousSelectedDashboardStates,
    state.selectedDashboard,
    state.selectedDashboardInputs,
  ]);
  /*eslint-enable */

  useEffect(() => {
    if (!state.availableDashboardsLoaded || !dashboardName) {
      return;
    }
    // If the dashboard we're viewing no longer exists, go back to the main page
    if (!state.dashboards.find((r) => r.full_name === dashboardName)) {
      navigate("/", { replace: true });
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
      type: "select_panel",
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
        dispatch,
        closePanelDetail,
      }}
    >
      <GlobalHotKeys
        allowChanges
        keyMap={hotKeysMap}
        handlers={hotKeysHandlers}
      />
      <FullHeightThemeWrapper>{children}</FullHeightThemeWrapper>
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
