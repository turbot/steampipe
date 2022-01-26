import findPathDeep from "deepdash/findPathDeep";
import { CheckLeafNodeExecutionTree } from "../components/reports/check/common";
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
import { get, set } from "lodash";
import { GlobalHotKeys } from "react-hotkeys";
import { LeafNodeData } from "../components/reports/common";
import { noop } from "../utils/func";
import { SOCKET_SERVER_URL } from "../config";
import { useNavigate, useParams } from "react-router-dom";

interface IReportContext {
  availableReportsLoaded: boolean;
  closePanelDetail(): void;
  dispatch(DispatchAction): void;
  error: any;
  reports: AvailableReport[];
  report: ReportDefinition | null;
  selectedPanel: PanelDefinition | null;
  selectedReport: AvailableReport | null;
}

interface AvailableReport {
  name: string;
  title: string;
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

export interface ReportDefinition {
  name: string;
  title?: string;
  width?: number;
  children?: (ContainerDefinition | PanelDefinition)[];
}

interface AvailableReportsDictionary {
  [key: string]: string;
}

interface SocketMessagePayload {
  action: string;
}

interface SocketMessage {
  data: string;
}

const ReportContext = createContext<IReportContext | null>(null);

const buildReportsList = (
  reports: AvailableReportsDictionary
): AvailableReport[] => {
  return Object.entries(reports).map(([name, title]) => ({
    name,
    title,
  }));
};

const updateSelectedReport = (
  selectedReport: AvailableReport,
  newReports: AvailableReport[]
) => {
  if (!selectedReport) {
    return null;
  }
  const matchingReport = newReports.find(
    (report) => report.name === selectedReport.name
  );
  if (matchingReport) {
    return matchingReport;
  } else {
    return null;
  }
};

function reducer(state, action) {
  switch (action.type) {
    case "available_reports":
      const reports = buildReportsList(action.reports);
      const selectedReport = updateSelectedReport(
        state.selectedReport,
        reports
      );
      return {
        ...state,
        availableReportsLoaded: true,
        reports,
        selectedReport: updateSelectedReport(state.selectedReport, reports),
        report:
          selectedReport && state.report.name === selectedReport.name
            ? state.report
            : null,
      };
    case "execution_started":
      // console.log("execution_started", { action, state });
      if (
        state.state === "complete" &&
        get(state, "report.name") === action.report_node.name
      ) {
        // console.log("Ignoring report execution started event", {
        //   action,
        //   state,
        // });
        return state;
      }
      return { ...state, error: null, report: action.report_node };
    case "execution_complete":
      // Replace the whole report as this event contains everything
      return {
        ...state,
        error: null,
        report: action.report_node,
        state: "complete",
      };
    case "leaf_node_progress":
    case "leaf_node_complete": {
      // Find the path to the name key that matches this panel and replace it
      const { report_node } = action;
      let panelPath: string = findPathDeep(
        state.report,
        (v, k) => k === "name" && v === report_node.name
      );

      if (!panelPath) {
        console.warn("Cannot find report panel to update", report_node.name);
        return state;
      }

      panelPath = panelPath.replace(".name", "");
      let newReport = {
        ...state.report,
      };
      newReport = set(newReport, panelPath, report_node);

      return {
        ...state,
        report: newReport,
      };
    }
    case "report_updated":
      return { ...state, report: action.report };
    case "select_panel":
      return { ...state, selectedPanel: action.panel };
    case "select_report":
      return {
        ...state,
        report: null,
        selectedReport: action.report,
        selectedPanel: null,
      };
    case "workspace_error":
      return { ...state, error: action.error };
    // Not emitting these from the report server yet
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

const ReportProvider = ({ children }) => {
  // const [reportState, send] = useStateMachine()({
  //   initial: "ready",
  //   states: {
  //     ready: {},
  //     started: {},
  //     complete: {},
  //     error: {},
  //   },
  //   verbose: true,
  // });
  const [state, dispatch] = useReducer(reducer, {
    reports: [],
    report: null,
    selectedPanel: null,
    selectedReport: null,
  });

  const { reportName } = useParams();
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
          // console.log("Sending keep alive");
          webSocket.current.send(JSON.stringify({ action: "keep_alive" }));
        }
        keepAliveTimerId = setTimeout(keepAlive, timeout);
      };

      if (!webSocket.current) {
        return;
      }

      // Send message to ask for available reports
      webSocket.current.send(
        JSON.stringify({
          action: "available_reports",
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

  // useEffect(() => {
  //   if (!webSocket.current) return;
  //   webSocket.current.send(
  //     JSON.stringify({
  //       action: "select_workspace",
  //       payload: { workspace: state.selectedWorkspace },
  //     })
  //   );
  // }, [state.selectedWorkspace]);

  useEffect(() => {
    if (!webSocket.current) return;

    if (!state.selectedReport || !state.selectedReport.name) return;

    webSocket.current.send(
      JSON.stringify({
        action: "select_report",
        payload: {
          // workspace: state.selectedWorkspace,
          report: {
            full_name: state.selectedReport.name,
          },
        },
      })
    );
  }, [state.selectedReport]);

  useEffect(() => {
    if (state.selectedReport && reportName === state.selectedReport.name) {
      return;
    }

    if (state.reports.length === 0) {
      return;
    }
    const report = state.reports.find((report) => report.name === reportName);
    dispatch({ type: "select_report", report });
  }, [reportName, state.selectedReport, state.reports]);

  useEffect(() => {
    if (!state.availableReportsLoaded || !reportName) {
      return;
    }
    // If the report we're viewing no longer exists, go back to the main page
    if (!state.reports.find((r) => r.name === reportName)) {
      navigate("/", { replace: true });
    }
  }, [reportName, state.availableReportsLoaded, state.reports]);

  useEffect(() => {
    if (!state.selectedReport) {
      document.title = "Reports | Steampipe";
    } else {
      document.title = `${
        state.selectedReport.title || state.selectedReport.name
      } | Reports | Steampipe`;
    }
  }, [state.selectedReport]);

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
  }, [state]);

  useEffect(() => {
    setHotKeysHandlers({
      CLOSE_PANEL_DETAIL: closePanelDetail,
    });
  }, [closePanelDetail]);

  return (
    <ReportContext.Provider value={{ ...state, dispatch, closePanelDetail }}>
      <GlobalHotKeys
        allowChanges
        keyMap={hotKeysMap}
        handlers={hotKeysHandlers}
      />
      <FullHeightThemeWrapper>{children}</FullHeightThemeWrapper>
    </ReportContext.Provider>
  );
};

const useReport = () => {
  const context = useContext(ReportContext);
  if (context === undefined) {
    throw new Error("useReport must be used within a ReportContext");
  }
  return context as IReportContext;
};

export { ReportContext, ReportProvider, useReport };
