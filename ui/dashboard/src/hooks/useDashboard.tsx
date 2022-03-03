// import * as AsBind from "as-bind";
import findPathDeep from "deepdash/findPathDeep";
import paths from "deepdash/paths";
import usePrevious from "./usePrevious";
import { CheckLeafNodeExecutionTree } from "../components/dashboards/check/common";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useReducer,
  useRef,
  useState,
} from "react";
import { FullHeightThemeWrapper } from "./useTheme";
import { get, isEqual, set, sortBy } from "lodash";
import { GlobalHotKeys } from "react-hotkeys";
import { LeafNodeData } from "../components/dashboards/common";
import { noop } from "../utils/func";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";

interface IDashboardContext {
  metadata: DashboardMetadata;
  metadataLoaded: boolean;
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

interface DashboardMetadata {
  mod: ModDashboardMetadata;
  installed_mods: InstalledModsDashboardMetadata;
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
  original_title?: string;
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

interface SocketMessagePayload {
  action: string;
}

interface SocketMessage {
  data: string;
}

const DashboardContext = createContext<IDashboardContext | null>(null);

const getSocketServerUrl = () => {
  // In this scenario the browser will be at http://localhost:3000,
  // so I have no idea what host + port the dashboard server is on
  if (process.env.NODE_ENV === "development") {
    return "ws://localhost:9194/ws";
  }
  // Otherwise, it's a production build, so use the URL details
  const url = new URL(window.location.toString());
  return `ws://${url.host}/ws`;
};

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
        metadataLoaded: true,
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

const initialiseInputs = (
  initialState: IDashboardContext,
  searchParams: URLSearchParams
) => ({
  ...initialState,
  selectedDashboardInputs:
    buildSelectedDashboardInputsFromSearchParams(searchParams),
});

const DashboardProvider = ({ children }) => {
  const [searchParams, setSearchParams] = useSearchParams();
  const [state, dispatch] = useReducer(
    reducer,
    {
      dashboards: [],
      dashboard: null,
      selectedPanel: null,
      selectedDashboard: null,
      selectedDashboardInputs: {},
      sqlDataMap: {},
    },
    (initialState) => initialiseInputs(initialState, searchParams)
  );

  const { dashboardName } = useParams();
  const navigate = useNavigate();
  const webSocket = useRef<WebSocket | null>(null);

  // useEffect(() => {
  //   const loadJqWasm = async () => {
  //     // @ts-ignore
  //     const jq = await fetch("../jq.wasm.wasm");
  //     // @ts-ignore
  //     const instance = await AsBind.instantiate(jq, {});
  //     console.log(jq);
  //     console.log(instance);
  //     // console.log(jq.json({ row: { name: "mike" } }, ".row.name"));
  //     // console.log(jq.json({ row: { name: "mike" } }, ".row.name"));
  //     // console.log(jq.json({ row: { name: "mike" } }, ".row.name"));
  //     // console.log(jq.json({ row: { name: "mike" } }, ".row.name"));
  //     // jq.json
  //   };
  //   loadJqWasm();
  // }, []);

  const onSocketError = (evt: any) => {
    console.error(evt);
  };

  const onSocketMessage = (message: SocketMessage) => {
    const payload: SocketMessagePayload = JSON.parse(message.data);
    dispatch({ type: payload.action, ...payload });
  };

  useEffect(() => {
    let keepAliveTimerId: NodeJS.Timeout;
    webSocket.current = new WebSocket(getSocketServerUrl());
    webSocket.current.onerror = onSocketError;
    webSocket.current.onmessage = onSocketMessage;
    webSocket.current.onopen = () => {
      const keepAlive = () => {
        if (!webSocket.current) {
          return;
        }

        const timeout = 20000;
        if (webSocket.current.readyState === webSocket.current.CLOSED) {
          webSocket.current = new WebSocket(getSocketServerUrl());
          webSocket.current.onerror = onSocketError;
          webSocket.current.onmessage = onSocketMessage;
        }
        if (webSocket.current.readyState === webSocket.current.OPEN) {
          webSocket.current.send(JSON.stringify({ action: "keep_alive" }));
        }
        keepAliveTimerId = setTimeout(keepAlive, timeout);
      };

      if (!webSocket.current) {
        return;
      }

      // Send message to ask for dashboard metadata
      webSocket.current.send(
        JSON.stringify({
          action: "get_dashboard_metadata",
        })
      );

      // Send message to ask for available dashboards
      webSocket.current.send(
        JSON.stringify({
          action: "get_available_dashboards",
        })
      );
      keepAlive();
    };
    return () => {
      clearTimeout(keepAliveTimerId);
      webSocket.current && webSocket.current.close();
    };
  }, []);

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
      console.log("Reinitialising dashboard inputs", value);
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
    if (
      !webSocket.current ||
      webSocket.current?.readyState !== webSocket.current.OPEN ||
      !state.selectedDashboard
    ) {
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
      webSocket.current.send(
        JSON.stringify({
          action: "clear_dashboard",
        })
      );
      webSocket.current.send(
        JSON.stringify({
          action: "select_dashboard",
          payload: {
            dashboard: {
              full_name: state.selectedDashboard.full_name,
            },
            input_values: state.selectedDashboardInputs,
          },
        })
      );
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
      webSocket.current.send(
        JSON.stringify({
          action: "input_changed",
          payload: {
            dashboard: {
              full_name: state.selectedDashboard.full_name,
            },
            changed_input: state.lastChangedInput,
            input_values: state.selectedDashboardInputs,
          },
        })
      );
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
    if (
      !webSocket.current ||
      webSocket.current?.readyState !== webSocket.current.OPEN ||
      state.selectedDashboard
    ) {
      return;
    }

    // If we've gone from having a report selected, to having nothing selected, clear the dashboard state
    if (previousSelectedDashboardStates)
      if (
        previousSelectedDashboardStates &&
        // @ts-ignore
        previousSelectedDashboardStates.selectedDashboard
      ) {
        webSocket.current.send(
          JSON.stringify({
            action: "clear_dashboard",
          })
        );
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
    // Sync params into the URL
    setSearchParams(state.selectedDashboardInputs);
  }, [previousSelectedDashboardStates, state.selectedDashboardInputs]);
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
    <DashboardContext.Provider value={{ ...state, dispatch, closePanelDetail }}>
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

export { DashboardContext, DashboardProvider, useDashboard };
