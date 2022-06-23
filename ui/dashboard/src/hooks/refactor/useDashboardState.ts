import isEqual from "lodash/isEqual";
import sortBy from "lodash/sortBy";
import usePrevious from "../usePrevious";
import {
  AvailableDashboard,
  AvailableDashboardsDictionary,
  DashboardActions,
  DashboardDataMode,
  DashboardInputs,
  DashboardsCollection,
  SelectedDashboardStates,
} from "../../types/dashboard";
import {
  addDataToPanels,
  buildSqlDataMap,
  calculateProgress,
  updatePanelsMapWithControlEvent,
  wrapDefinitionInArtificialDashboard,
} from "../../utils/dashboard";
import { SocketActions } from "../../types/webSocket";
import { useEffect, useReducer } from "react";
import { useNavigate, useParams } from "react-router-dom";

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

const dashboardReducer = (state, action, context) => {
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
        selectedDashboard:
          context.dataMode === "snapshot"
            ? state.selectedDashboard
            : selectedDashboard,
        dashboard:
          state.dataMode === "snapshot"
            ? state.dashboard
            : selectedDashboard &&
              state.dashboard &&
              state.dashboard.name === selectedDashboard.full_name
            ? state.dashboard
            : null,
      };
    case DashboardActions.SET_DASHBOARD_TAG_KEYS:
      return {
        ...state,
        dashboardTags: {
          ...state.dashboardTags,
          keys: action.keys,
        },
      };
    case DashboardActions.SELECT_DASHBOARD:
      if (context.dataMode === "snapshot") {
        return {
          ...state,
          error: null,
          selectedDashboard: action.dashboard,
        };
      }
      return {
        ...state,
        dashboard: null,
        error: null,
        execution_id: null,
        snapshotId: null,
        state: null,
        selectedDashboard: action.dashboard,
        selectedPanel: null,
        lastChangedInput: null,
      };
    case DashboardActions.SELECT_PANEL:
      return { ...state, selectedPanel: action.panel };
    case DashboardActions.EXECUTION_STARTED: {
      const originalDashboard = action.dashboard_node;
      let dashboard;
      // For benchmarks and controls that are run directly from a mod,
      // we need to wrap these in an artificial dashboard, so we can treat
      // it just like any other dashboard
      if (action.dashboard_node.panel_type !== "dashboard") {
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
        panelsMap: addDataToPanels(action.panels, state.sqlDataMap),
        dashboard,
        execution_id: action.execution_id,
        refetchDashboard: false,
        progress: 0,
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

      if (action.dashboard_node.panel_type !== "dashboard") {
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
        progress: 100,
        state: "complete",
      };
    }
    case DashboardActions.EXECUTION_ERROR:
      return { ...state, error: action.error, progress: 100, state: "error" };
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
        progress: calculateProgress(updatedPanelsMap),
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
        progress: calculateProgress(panelsMap),
      };
    }
    case DashboardActions.WORKSPACE_ERROR:
      return { ...state, error: action.error };
    default:
      return state;
  }
};

const getInitialDashboardState = () => ({
  availableDashboardsLoaded: false,
  dashboardTags: {
    keys: [],
  },
  dashboard: null,
  dashboards: [],
  dashboardsMap: {},
  panelsMap: {},
  selectedDashboard: null,
  sqlDataMap: {},
});

const useDashboardState = (
  inputs: DashboardInputs,
  dataMode: DashboardDataMode,
  analyticsContext: any,
  socketReady: boolean,
  sendSocketMessage: (message: any) => void
) => {
  const navigate = useNavigate();
  const [dashboardState, dispatch] = useReducer(
    (action, state) => dashboardReducer(action, state, { dataMode }),
    getInitialDashboardState()
  );
  const { dashboard_name } = useParams();

  // Keep track of the previous selected dashboard and inputs
  const previousSelectedDashboardStates: SelectedDashboardStates | undefined =
    usePrevious({
      dashboard_name,
      selectedDashboard: dashboardState.selectedDashboard,
      inputs,
    });

  useEffect(() => {
    analyticsContext.setMetadata(dashboardState.metadata);
  }, [dashboardState.metadata, analyticsContext.setMetadata]);

  useEffect(() => {
    analyticsContext.setSelectedDashboard(dashboardState.selectedDashboard);
  }, [dashboardState.selectedDashboard, analyticsContext.setSelectedDashboard]);

  useEffect(() => {
    // If we've got no dashboard selected in the URL, but we've got one selected in state,
    // then clear both the inputs and the selected dashboard in state
    if (!dashboard_name && dashboardState.selectedDashboard) {
      // dispatch({
      //   type: DashboardActions.CLEAR_DASHBOARD_INPUTS,
      //   recordInputsHistory: false,
      // });
      dispatch({
        type: DashboardActions.SELECT_DASHBOARD,
        dashboard: null,
        recordInputsHistory: false,
      });
      return;
    }
    // Else if we've got a dashboard selected in the URL and don't have one selected in state,
    // select that dashboard
    if (dashboard_name && !dashboardState.selectedDashboard) {
      const dashboard = dashboardState.dashboards.find(
        (dashboard) => dashboard.full_name === dashboard_name
      );
      dispatch({
        type: DashboardActions.SELECT_DASHBOARD,
        dashboard,
        dataMode: dataMode,
      });
      return;
    }
    // Else if we've changed to a different report in the URL then clear the inputs and select the
    // dashboard in state
    if (
      dashboard_name &&
      dashboardState.selectedDashboard &&
      dashboard_name !== dashboardState.selectedDashboard.full_name
    ) {
      const dashboard = dashboardState.dashboards.find(
        (dashboard) => dashboard.full_name === dashboard_name
      );
      dispatch({ type: DashboardActions.SELECT_DASHBOARD, dashboard });
    }
  }, [
    dashboard_name,
    dispatch,
    dashboardState.dashboards,
    dataMode,
    dashboardState.selectedDashboard,
  ]);

  useEffect(() => {
    if (!dashboardState.availableDashboardsLoaded || !dashboard_name) {
      return;
    }
    // If the dashboard we're viewing no longer exists, go back to the main page
    if (
      !dashboardState.dashboards.find((r) => r.full_name === dashboard_name)
    ) {
      navigate("../", { replace: true });
    }
  }, [
    navigate,
    dashboard_name,
    dashboardState.availableDashboardsLoaded,
    dashboardState.dashboards,
  ]);

  useEffect(() => {
    // This effect will send events over websockets and depends on there being a dashboard selected
    if (!socketReady || !dashboardState.selectedDashboard) {
      return;
    }

    // If we didn't previously have a dashboard selected in state (e.g. you've gone from home page
    // to a report, or it's first load), or the selected dashboard has been changed, select that
    // report over the socket
    if (
      dataMode === "live" &&
      (!previousSelectedDashboardStates ||
        // @ts-ignore
        !previousSelectedDashboardStates.selectedDashboard ||
        dashboardState.selectedDashboard.full_name !==
          // @ts-ignore
          previousSelectedDashboardStates.selectedDashboard.full_name)
    ) {
      sendSocketMessage({
        action: SocketActions.CLEAR_DASHBOARD,
      });
      sendSocketMessage({
        action: SocketActions.SELECT_DASHBOARD,
        payload: {
          dashboard: {
            full_name: dashboardState.selectedDashboard.full_name,
          },
          input_values: inputs,
        },
      });
      return;
    }
    // Else if we did previously have a dashboard selected in state and the
    // inputs have changed, then update the inputs over the socket
    if (
      dataMode === "live" &&
      previousSelectedDashboardStates &&
      // @ts-ignore
      previousSelectedDashboardStates.selectedDashboard &&
      !isEqual(
        // @ts-ignore
        previousSelectedDashboardStates.inputs,
        inputs
      )
    ) {
      sendSocketMessage({
        action: SocketActions.INPUT_CHANGED,
        payload: {
          dashboard: {
            full_name: dashboardState.selectedDashboard.full_name,
          },
          // changed_input: state.lastChangedInput,
          input_values: inputs,
        },
      });
    }
  }, [
    previousSelectedDashboardStates,
    sendSocketMessage,
    socketReady,
    dashboardState.selectedDashboard,
    inputs,
    dataMode,
  ]);

  useEffect(() => {
    // This effect will send events over websockets and depends on there being no dashboard selected
    if (!socketReady || dashboardState.selectedDashboard) {
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
    dashboardState.selectedDashboard,
  ]);

  return {
    state: dashboardState,
    dispatch,
  };
};

export default useDashboardState;
