import get from "lodash/get";
import useDashboardVersionCheck from "./useDashboardVersionCheck";
import {
  buildDashboards,
  buildPanelsLog,
  buildSelectedDashboardInputsFromSearchParams,
  updatePanelsLogFromCompletedPanels,
  updateSelectedDashboard,
  wrapDefinitionInArtificialDashboard,
} from "../utils/state";
import {
  controlsUpdatedEventHandler,
  leafNodesUpdatedEventHandler,
  migratePanelStatuses,
} from "../utils/dashboardEventHandlers";
import {
  DashboardActions,
  DashboardDataModeCLISnapshot,
  DashboardDataModeCloudSnapshot,
  DashboardDataModeLive,
  DashboardDataOptions,
  DashboardRenderOptions,
  IDashboardContext,
} from "../types";
import {
  EXECUTION_SCHEMA_VERSION_20220929,
  EXECUTION_SCHEMA_VERSION_20221222,
} from "../constants/versions";
import {
  ExecutionCompleteSchemaMigrator,
  ExecutionStartedSchemaMigrator,
} from "../utils/schema";
import { useCallback, useReducer } from "react";

const reducer = (state: IDashboardContext, action) => {
  switch (action.type) {
    case DashboardActions.DASHBOARD_METADATA:
      return {
        ...state,
        metadata: {
          mod: {},
          ...action.metadata,
        },
      };
    case DashboardActions.AVAILABLE_DASHBOARDS:
      const { dashboards, dashboardsMap } = buildDashboards(
        action.dashboards,
        action.benchmarks,
        action.snapshots,
      );
      const selectedDashboard = updateSelectedDashboard(
        state.selectedDashboard,
        dashboards,
      );
      return {
        ...state,
        error: null,
        availableDashboardsLoaded: true,
        dashboards,
        dashboardsMap,
        selectedDashboard:
          state.dataMode === DashboardDataModeCLISnapshot ||
          state.dataMode === DashboardDataModeCloudSnapshot
            ? state.selectedDashboard
            : selectedDashboard,
        dashboard:
          state.dataMode === DashboardDataModeCLISnapshot ||
          state.dataMode === DashboardDataModeCloudSnapshot
            ? state.dashboard
            : selectedDashboard &&
              state.dashboard &&
              state.dashboard.name === selectedDashboard.full_name
            ? state.dashboard
            : null,
      };
    case DashboardActions.EXECUTION_STARTED: {
      const rootLayoutPanel = action.layout;
      const rootPanel = action.panels[rootLayoutPanel.name];
      let dashboard;
      // For benchmarks and controls that are run directly from a mod,
      // we need to wrap these in an artificial dashboard, so we can treat
      // it just like any other dashboard
      if (rootPanel.panel_type !== "dashboard") {
        dashboard = wrapDefinitionInArtificialDashboard(
          rootPanel,
          action.layout,
        );
      } else {
        dashboard = {
          ...rootPanel,
          ...action.layout,
        };
      }

      const eventMigrator = new ExecutionStartedSchemaMigrator();
      const migratedEvent = eventMigrator.toLatest(action);

      return {
        ...state,
        error: null,
        panelsLog: buildPanelsLog(
          migratedEvent.panels,
          migratedEvent.start_time,
        ),
        panelsMap: migratedEvent.panels,
        diff: null,
        dashboard,
        execution_id: migratedEvent.execution_id,
        refetchDashboard: false,
        progress: 0,
        snapshot: null,
        state: "running",
      };
    }
    case DashboardActions.DIFF_SNAPSHOT: {
      // If we're in live mode, do nothing
      if (state.dataMode === DashboardDataModeLive) {
        return state;
      }

      const eventMigrator = new ExecutionCompleteSchemaMigrator();
      const migratedEvent = eventMigrator.toLatest(action);
      const panels = migratedEvent.snapshot.panels;
      const panelsMap = migratePanelStatuses(panels, action.schema_version);

      return {
        ...state,
        diff: {
          panelsMap,
          snapshotFileName: action.snapshotFileName,
        },
      };
    }
    case DashboardActions.EXECUTION_COMPLETE: {
      // If we're in live mode and not expecting execution events for this ID
      if (
        state.dataMode === DashboardDataModeLive &&
        action.execution_id !== state.execution_id
      ) {
        return state;
      }

      const eventMigrator = new ExecutionCompleteSchemaMigrator();
      const migratedEvent = eventMigrator.toLatest(action);
      const layout = migratedEvent.snapshot.layout;
      const panels = migratedEvent.snapshot.panels;
      const rootLayoutPanel = migratedEvent.snapshot.layout;
      const rootPanel = panels[rootLayoutPanel.name];
      let dashboard;

      if (rootPanel.panel_type !== "dashboard") {
        dashboard = wrapDefinitionInArtificialDashboard(rootPanel, layout);
      } else {
        dashboard = {
          ...rootPanel,
          ...layout,
        };
      }

      const panelsMap = migratePanelStatuses(panels, action.schema_version);

      // Replace the whole dashboard as this event contains everything
      return {
        ...state,
        error: null,
        panelsLog: updatePanelsLogFromCompletedPanels(
          state.panelsLog,
          panels,
          action.snapshot.end_time,
        ),
        diff: null,
        panelsMap,
        dashboard,
        progress: 100,
        snapshot: action.snapshot,
        state: "complete",
      };
    }
    case DashboardActions.EXECUTION_ERROR:
      return { ...state, error: action.error, progress: 100, state: "error" };
    case DashboardActions.CONTROLS_UPDATED:
      return controlsUpdatedEventHandler(action, state);
    case DashboardActions.LEAF_NODES_COMPLETE:
      return leafNodesUpdatedEventHandler(
        action,
        EXECUTION_SCHEMA_VERSION_20220929,
        state,
      );
    case DashboardActions.LEAF_NODES_UPDATED:
      return leafNodesUpdatedEventHandler(
        action,
        EXECUTION_SCHEMA_VERSION_20221222,
        state,
      );
    case DashboardActions.SELECT_PANEL:
      return { ...state, selectedPanel: action.panel };
    case DashboardActions.SET_DATA_MODE:
      const newState = {
        ...state,
        dataMode: action.dataMode,
      };
      if (action.dataMode === DashboardDataModeCLISnapshot) {
        newState.snapshotFileName = action.snapshotFileName;
      } else if (
        state.dataMode !== DashboardDataModeLive &&
        action.dataMode === DashboardDataModeLive
      ) {
        newState.snapshot = null;
        newState.snapshotFileName = null;
        newState.snapshotId = null;
      }
      return newState;
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
      if (action.dashboard && action.dashboard.type === "snapshot") {
        return {
          ...state,
          dataMode: DashboardDataModeCLISnapshot,
          selectedDashboard: action.dashboard,
        };
      }

      if (
        action.dataMode === DashboardDataModeCLISnapshot ||
        action.dataMode === DashboardDataModeCloudSnapshot
      ) {
        return {
          ...state,
          dataMode: action.dataMode,
          selectedDashboard: action.dashboard,
        };
      }

      return {
        ...state,
        dataMode: DashboardDataModeLive,
        dashboard: null,
        execution_id: null,
        panelsMap: {},
        snapshot: null,
        snapshotFileName: null,
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
      const newPanelsMap = { ...state.panelsMap };
      const panelsMapKeys = Object.keys(newPanelsMap);
      for (const input of action.cleared_inputs || []) {
        delete newSelectedDashboardInputs[input];
        const matchingPanelKey = panelsMapKeys.find((key) =>
          key.endsWith(input),
        );
        if (!matchingPanelKey) {
          continue;
        }
        const panel = newPanelsMap[matchingPanelKey];
        newPanelsMap[matchingPanelKey] = {
          ...panel,
          status: "initialized",
        };
      }
      return {
        ...state,
        panelsMap: newPanelsMap,
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
};

const getInitialState = (searchParams, defaults: any = {}) => {
  return {
    versionMismatchCheck: defaults.versionMismatchCheck,
    availableDashboardsLoaded: false,
    metadata: null,
    dashboards: [],
    dashboardTags: {
      keys: [],
    },
    dataMode: defaults.dataMode || DashboardDataModeLive,
    snapshotId: defaults.snapshotId ? defaults.snapshotId : null,
    refetchDashboard: false,
    error: null,
    panelsLog: {},
    panelsMap: {},
    dashboard: null,
    selectedPanel: null,
    selectedDashboard: null,
    selectedDashboardInputs:
      buildSelectedDashboardInputsFromSearchParams(searchParams),
    snapshot: null,
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

    execution_id: null,

    progress: 0,
  };
};

type DashboardStateProps = {
  dataOptions: DashboardDataOptions;
  renderOptions: DashboardRenderOptions;
  searchParams: URLSearchParams;
  stateDefaults: {};
  versionMismatchCheck: boolean;
};

const useDashboardState = ({
  dataOptions = {
    dataMode: DashboardDataModeLive,
  },
  renderOptions = {
    headless: false,
  },
  searchParams,
  stateDefaults = {},
  versionMismatchCheck,
}: DashboardStateProps) => {
  const [state, dispatchInner] = useReducer(
    reducer,
    getInitialState(searchParams, {
      ...stateDefaults,
      ...dataOptions,
      ...renderOptions,
      versionMismatchCheck,
    }),
  );
  useDashboardVersionCheck(state);
  const dispatch = useCallback((action) => {
    // console.log(action.type, action);
    dispatchInner(action);
  }, []);
  return [state, dispatch];
};

export default useDashboardState;
