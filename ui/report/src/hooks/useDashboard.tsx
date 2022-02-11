import findPathDeep from "deepdash/findPathDeep";
import paths from "deepdash/paths";
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
import { get, set, sortBy } from "lodash";
import { GlobalHotKeys } from "react-hotkeys";
import { LeafNodeData } from "../components/dashboards/common";
import { noop } from "../utils/func";
import { SOCKET_SERVER_URL } from "../config";
import { useNavigate, useParams } from "react-router-dom";

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
  sqlDataMap: SQLDataMap;
}

interface ModDashboardMetadata {
  full_name: string;
  short_name: string;
}

interface DashboardMetadata {
  mod: ModDashboardMetadata;
}

export interface AvailableDashboard {
  full_name: string;
  short_name: string;
  mod_full_name: string;
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
  execution_tree?: CheckLeafNodeExecutionTree;
  error?: Error;
  properties?: PanelProperties;
}

export interface DashboardDefinition {
  name: string;
  title?: string;
  width?: number;
  children?: (ContainerDefinition | PanelDefinition)[];
}

interface SocketMessagePayload {
  action: string;
}

interface SocketMessage {
  data: string;
}

const DashboardContext = createContext<IDashboardContext | null>(null);

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
  // const justSQL = pickDeep(dashboard, ["sql"]);
  // console.log(justSQL);
  const sqlPaths = paths(dashboard, { leavesOnly: true }).filter((path) =>
    path.endsWith(".sql")
  );
  // console.log(sqlPaths);
  const sqlDataMap = {};
  for (const sqlPath of sqlPaths) {
    const sql = get(dashboard, sqlPath);
    // console.log(dashboard, sql);
    const dataPath = `${sqlPath.substring(0, sqlPath.indexOf(".sql"))}.data`;
    const data = get(dashboard, dataPath);
    if (!sqlDataMap[sql]) {
      sqlDataMap[sql] = data;
    }
  }
  // console.log(sqlDataMap);
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
      // console.log("execution_started", { action, state });
      // if (
      //   state.state === "complete" &&
      //   get(state, "dashboard.name") === action.report_node.name
      // ) {
      //   // console.log("Ignoring dashboard execution started event", {
      //   //   action,
      //   //   state,
      //   // });
      //   return state;
      // }
      // console.log("Started", action.report_node);
      const dashboardWithData = addDataToDashboard(
        action.report_node,
        state.sqlDataMap
      );
      return {
        ...state,
        error: null,
        dashboard: dashboardWithData,
        state: "running",
      };
    case "execution_complete":
      // console.log("Complete", action.report_node);
      // Build map of SQL to data
      const sqlDataMap = buildSqlDataMap(action.report_node);
      // console.log(sqlDataMap);
      // Replace the whole dashboard as this event contains everything
      return {
        ...state,
        error: null,
        dashboard: action.report_node,
        sqlDataMap,
        state: "complete",
      };
    case "leaf_node_progress":
    case "leaf_node_complete": {
      // Find the path to the name key that matches this panel and replace it
      const { report_node } = action;
      let panelPath: string = findPathDeep(
        state.dashboard,
        (v, k) => k === "name" && v === report_node.name
      );

      if (!panelPath) {
        console.warn("Cannot find dashboard panel to update", report_node.name);
        return state;
      }

      panelPath = panelPath.replace(".name", "");
      let newDashboard = {
        ...state.dashboard,
      };
      newDashboard = set(newDashboard, panelPath, report_node);

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

const DashboardProvider = ({ children }) => {
  const [state, dispatch] = useReducer(reducer, {
    dashboards: [],
    dashboard: null,
    selectedPanel: null,
    selectedDashboard: null,
    sqlDataMap: {},
  });

  const { dashboardName } = useParams();
  const navigate = useNavigate();
  const webSocket = useRef<WebSocket | null>(null);

  const onSocketError = (evt: any) => {
    console.error(evt);
  };

  const onSocketMessage = (message: SocketMessage) => {
    const payload: SocketMessagePayload = JSON.parse(message.data);
    // console.log({ message, payload });
    dispatch({ type: payload.action, ...payload });
  };

  useEffect(() => {
    let keepAliveTimerId: NodeJS.Timeout;
    webSocket.current = new WebSocket(SOCKET_SERVER_URL);
    webSocket.current.onerror = onSocketError;
    webSocket.current.onmessage = onSocketMessage;
    webSocket.current.onopen = () => {
      const keepAlive = () => {
        if (!webSocket.current) {
          return;
        }

        const timeout = 20000;
        // console.log("Trying to send keep alive", webSocket.current.readyState);
        if (webSocket.current.readyState === webSocket.current.CLOSED) {
          // console.log("Socket closed. Re-opening");
          webSocket.current = new WebSocket(SOCKET_SERVER_URL);
          webSocket.current.onerror = onSocketError;
          webSocket.current.onmessage = onSocketMessage;
        }
        if (webSocket.current.readyState === webSocket.current.OPEN) {
          // console.log("Sending keep alive", webSocket.current);
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
      // console.log("Clearing keep alive and closing socket");
      clearTimeout(keepAliveTimerId);
      webSocket.current && webSocket.current.close();
    };
  }, []);

  useEffect(() => {
    if (
      !webSocket.current ||
      webSocket.current?.readyState !== webSocket.current.OPEN
    ) {
      return;
    }

    if (!state.selectedDashboard || !state.selectedDashboard.full_name) {
      return;
    }

    webSocket.current.send(
      JSON.stringify({
        action: "select_dashboard",
        payload: {
          // workspace: state.selectedWorkspace,
          dashboard: {
            full_name: state.selectedDashboard.full_name,
          },
        },
      })
    );
  }, [state.selectedDashboard]);

  useEffect(() => {
    if (
      state.selectedDashboard &&
      dashboardName === state.selectedDashboard.full_name
    ) {
      return;
    }

    if (state.dashboards.length === 0) {
      return;
    }
    const dashboard = state.dashboards.find(
      (dashboard) => dashboard.full_name === dashboardName
    );
    dispatch({ type: "select_dashboard", dashboard });
  }, [dashboardName, state.selectedDashboard, state.dashboards]);

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
