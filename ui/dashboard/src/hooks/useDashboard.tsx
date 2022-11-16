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
import useDashboardWebSocketEventHandler from "./useDashboardWebSocketEventHandler";
import usePrevious from "./usePrevious";
import VersionErrorMismatch from "../components/VersionErrorMismatch";
import {
  AvailableDashboard,
  AvailableDashboardsDictionary,
  DashboardActions,
  DashboardDataModeCLISnapshot,
  DashboardDataModeCloudSnapshot,
  DashboardDataModeLive,
  DashboardDataOptions,
  DashboardDefinition,
  DashboardRenderOptions,
  DashboardsCollection,
  IDashboardContext,
  PanelDefinition,
  PanelsMap,
  SelectedDashboardStates,
  SocketURLFactory,
  SQLDataMap,
} from "../types";
import { buildComponentsMap } from "../components";
import {
  controlsUpdatedEventHandler,
  leafNodesCompleteEventHandler,
  migrateDashboardExecutionCompleteSchema,
} from "../utils/dashboardEventHandlers";
import { GlobalHotKeys } from "react-hotkeys";
import { KeyValueStringPairs } from "../components/dashboards/common/types";
import { noop } from "../utils/func";
import {
  useLocation,
  useNavigate,
  useNavigationType,
  useParams,
  useSearchParams,
} from "react-router-dom";

const buildDashboards = (
  dashboards: AvailableDashboardsDictionary,
  benchmarks: AvailableDashboardsDictionary,
  snapshots: KeyValueStringPairs
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

  for (const snapshot of Object.keys(snapshots || {})) {
    const builtSnapshot: AvailableDashboard = {
      title: snapshot,
      full_name: snapshot,
      short_name: snapshot,
      type: "snapshot",
      tags: {},
      is_top_level: true,
    };
    dashboardsMap[builtSnapshot.full_name] = builtSnapshot;
    builtDashboards.push(builtSnapshot);
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
    const panelPath = `${sqlPath.substring(0, sqlPath.indexOf(".sql"))}`;
    const dataPath = `${panelPath}.data`;
    const panel = get(panels, panelPath);
    // We don't want to retain panel data for inputs as it causes issues with selection
    // of incorrect values for select controls without placeholders
    if (panel && panel.panel_type !== "input") {
      set(panels, dataPath, data);
    }
  }
  return { ...panels };
}

const wrapDefinitionInArtificialDashboard = (
  definition: PanelDefinition,
  layout: any
): DashboardDefinition => {
  const { title: defTitle, ...definitionWithoutTitle } = definition;
  const { title: layoutTitle, ...layoutWithoutTitle } = layout;
  return {
    artificial: true,
    name: definition.name,
    title: definition.title,
    panel_type: "dashboard",
    children: [
      {
        ...definitionWithoutTitle,
        ...layoutWithoutTitle,
      },
    ],
    dashboard: definition.dashboard,
  };
};

function reducer(state, action) {
  if (state.ignoreEvents) {
    return state;
  }

  switch (action.type) {
    case DashboardActions.DASHBOARD_METADATA:
      let cliVersion: string | null = "";
      let uiVersion: string | null = "";
      let mismatchedVersions = false;
      if (state.versionMismatchCheck) {
        const cliVersionRaw = get(action.metadata, "cli.version");
        const uiVersionRaw = process.env.REACT_APP_VERSION;
        const hasVersionsSet = !!cliVersionRaw && !!uiVersionRaw;
        cliVersion = !!cliVersionRaw
          ? cliVersionRaw.startsWith("v")
            ? cliVersionRaw.substring(1)
            : cliVersionRaw
          : null;
        uiVersion = !!uiVersionRaw
          ? uiVersionRaw.startsWith("v")
            ? uiVersionRaw.substring(1)
            : uiVersionRaw
          : null;
        mismatchedVersions = hasVersionsSet && cliVersion !== uiVersion;
      }
      return {
        ...state,
        metadata: {
          mod: {},
          ...action.metadata,
        },
        error: mismatchedVersions ? (
          <VersionErrorMismatch cliVersion={cliVersion} uiVersion={uiVersion} />
        ) : null,
        ignoreEvents: mismatchedVersions,
      };
    case DashboardActions.AVAILABLE_DASHBOARDS:
      const { dashboards, dashboardsMap } = buildDashboards(
        action.dashboards,
        action.benchmarks,
        action.snapshots
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
          action.layout
        );
      } else {
        dashboard = {
          ...rootPanel,
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
        snapshot: null,
        state: "ready",
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

      const migratedEvent = migrateDashboardExecutionCompleteSchema(action);
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

      // Build map of SQL to data
      const sqlDataMap = buildSqlDataMap(panels);
      // Replace the whole dashboard as this event contains everything
      return {
        ...state,
        error: null,
        panelsMap: panels,
        dashboard,
        sqlDataMap,
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
      return leafNodesCompleteEventHandler(action, state);
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
        sqlDataMap: {},
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
          key.endsWith(input)
        );
        if (!matchingPanelKey) {
          continue;
        }
        const panel = newPanelsMap[matchingPanelKey];
        newPanelsMap[matchingPanelKey] = {
          ...panel,
          status: "ready",
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

const getInitialState = (searchParams, defaults: any = {}) => {
  return {
    versionMismatchCheck: defaults.versionMismatchCheck,
    ignoreEvents: false,
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

    sqlDataMap: {},

    execution_id: null,

    progress: 0,
  };
};

const DashboardContext = createContext<IDashboardContext | null>(null);

interface DashboardProviderProps {
  analyticsContext: any;
  breakpointContext: any;
  children: null | JSX.Element | JSX.Element[];
  componentOverrides?: {};
  dataOptions?: DashboardDataOptions;
  eventHooks?: {};
  featureFlags?: string[];
  renderOptions?: DashboardRenderOptions;
  socketUrlFactory?: SocketURLFactory;
  stateDefaults?: {};
  themeContext: any;
  versionMismatchCheck?: boolean;
}

const DashboardProvider = ({
  analyticsContext,
  breakpointContext,
  children,
  componentOverrides = {},
  dataOptions = {
    dataMode: DashboardDataModeLive,
  },
  eventHooks,
  featureFlags = [],
  renderOptions = {
    headless: false,
  },
  socketUrlFactory,
  stateDefaults = {},
  versionMismatchCheck = false,
  themeContext,
}: DashboardProviderProps) => {
  const components = buildComponentsMap(componentOverrides);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [state, dispatchInner] = useReducer(
    reducer,
    getInitialState(searchParams, {
      ...stateDefaults,
      ...dataOptions,
      ...renderOptions,
      versionMismatchCheck,
    })
  );
  const dispatch = useCallback((action) => {
    // console.log(action.type, action);
    dispatchInner(action);
  }, []);
  const { dashboard_name } = useParams();
  const { eventHandler } = useDashboardWebSocketEventHandler(
    dispatch,
    eventHooks
  );
  const { ready: socketReady, send: sendSocketMessage } = useDashboardWebSocket(
    state.dataMode,
    dispatch,
    eventHandler,
    socketUrlFactory
  );
  const {
    setMetadata: setAnalyticsMetadata,
    setSelectedDashboard: setAnalyticsSelectedDashboard,
  } = analyticsContext;

  const location = useLocation();
  const navigationType = useNavigationType();

  // Keep track of the previous selected dashboard and inputs
  const previousSelectedDashboardStates: SelectedDashboardStates | undefined =
    usePrevious<SelectedDashboardStates>({
      dashboard_name,
      dataMode: state.dataMode,
      refetchDashboard: state.refetchDashboard,
      search: state.search,
      searchParams,
      selectedDashboard: state.selectedDashboard,
      selectedDashboardInputs: state.selectedDashboardInputs,
    });

  // Alert analytics
  useEffect(() => {
    setAnalyticsMetadata(state.metadata);
  }, [state.metadata, setAnalyticsMetadata]);

  useEffect(() => {
    setAnalyticsSelectedDashboard(state.selectedDashboard);
  }, [state.selectedDashboard, setAnalyticsSelectedDashboard]);

  useEffect(() => {
    if (
      !!dashboard_name &&
      !location.pathname.startsWith("/snapshot/") &&
      state.dataMode === DashboardDataModeCLISnapshot
    ) {
      dispatch({
        type: DashboardActions.SET_DATA_MODE,
        dataMode: DashboardDataModeLive,
      });
    }
  }, [dashboard_name, dispatch, location, navigate, state.dataMode]);

  // Ensure that on history pop / push we sync the new values into state
  useEffect(() => {
    if (navigationType !== "POP" && navigationType !== "PUSH") {
      return;
    }
    if (location.key === "default") {
      return;
    }
    if (state.dataMode !== DashboardDataModeLive) {
      return;
    }

    // If we've just popped or pushed from one dashboard to another, then we don't want to add the search to the URL
    // as that will show the dashboard list, but we want to see the dashboard that we came from / went to previously.
    const goneFromDashboardToDashboard =
      previousSelectedDashboardStates?.dashboard_name &&
      dashboard_name &&
      previousSelectedDashboardStates.dashboard_name !== dashboard_name;

    const search = searchParams.get("search") || "";
    const groupBy =
      searchParams.get("group_by") ||
      get(stateDefaults, "search.groupBy.value", "tag");
    const tag =
      searchParams.get("tag") ||
      get(stateDefaults, "search.groupBy.tag", "service");
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
    if (
      JSON.stringify(
        previousSelectedDashboardStates?.selectedDashboardInputs
      ) !== JSON.stringify(inputs)
    ) {
      dispatch({
        type: DashboardActions.SET_DASHBOARD_INPUTS,
        value: inputs,
        recordInputsHistory: false,
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
    state.dataMode,
  ]);

  useEffect(() => {
    // If no search params have changed
    if (
      state.dataMode === DashboardDataModeCloudSnapshot ||
      state.dataMode === DashboardDataModeCLISnapshot ||
      (previousSelectedDashboardStates &&
        previousSelectedDashboardStates?.dashboard_name === dashboard_name &&
        previousSelectedDashboardStates.dataMode === state.dataMode &&
        previousSelectedDashboardStates.search.value === state.search.value &&
        previousSelectedDashboardStates.search.groupBy.value ===
          state.search.groupBy.value &&
        previousSelectedDashboardStates.search.groupBy.tag ===
          state.search.groupBy.tag &&
        previousSelectedDashboardStates.searchParams.toString() ===
          searchParams.toString())
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
    if (
      dashboard_name &&
      !state.selectedDashboard &&
      state.dataMode === DashboardDataModeLive
    ) {
      const dashboard = state.dashboards.find(
        (dashboard) => dashboard.full_name === dashboard_name
      );
      dispatch({
        type: DashboardActions.SELECT_DASHBOARD,
        dashboard,
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
    if (
      !dashboard_name &&
      state.snapshot &&
      state.dataMode === DashboardDataModeCLISnapshot
    ) {
      dispatch({
        type: DashboardActions.SELECT_DASHBOARD,
        dashboard: null,
        dataMode: DashboardDataModeLive,
      });
    }
  }, [dashboard_name, dispatch, state.dataMode, state.snapshot]);

  useEffect(() => {
    // This effect will send events over websockets and depends on there being a dashboard selected
    if (!socketReady || !state.selectedDashboard) {
      return;
    }

    // If we didn't previously have a dashboard selected in state (e.g. you've gone from home page
    // to a report, or it's first load), or the selected dashboard has been changed, select that
    // report over the socket
    if (
      (state.dataMode === DashboardDataModeLive ||
        state.dataMode === DashboardDataModeCLISnapshot) &&
      (!previousSelectedDashboardStates ||
        !previousSelectedDashboardStates.selectedDashboard ||
        state.selectedDashboard.full_name !==
          previousSelectedDashboardStates.selectedDashboard.full_name ||
        (!previousSelectedDashboardStates.refetchDashboard &&
          state.refetchDashboard))
    ) {
      sendSocketMessage({
        action: SocketActions.CLEAR_DASHBOARD,
      });
      sendSocketMessage({
        action:
          state.selectedDashboard.type === "snapshot"
            ? SocketActions.SELECT_SNAPSHOT
            : SocketActions.SELECT_DASHBOARD,
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
      state.dataMode === DashboardDataModeLive &&
      previousSelectedDashboardStates &&
      previousSelectedDashboardStates.selectedDashboard &&
      !isEqual(
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
        previousSelectedDashboardStates.selectedDashboardInputs
      )
    ) {
      return;
    }

    // Only record history when it's the same report before and after and the inputs have changed
    const shouldRecordHistory =
      state.recordInputsHistory &&
      !!previousSelectedDashboardStates.selectedDashboard &&
      !!state.selectedDashboard &&
      previousSelectedDashboardStates.selectedDashboard.full_name ===
        state.selectedDashboard.full_name;

    // Sync params into the URL
    const newParams = {
      ...state.selectedDashboardInputs,
    };
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
    if (
      !state.availableDashboardsLoaded ||
      !dashboard_name ||
      state.dataMode === DashboardDataModeCLISnapshot
    ) {
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
    state.dataMode,
  ]);

  useEffect(() => {
    if (
      location.pathname.startsWith("/snapshot/") &&
      state.dataMode !== DashboardDataModeCLISnapshot
    ) {
      navigate("/");
    }
  }, [location, navigate, state.dataMode]);

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

  const [renderSnapshotCompleteDiv, setRenderSnapshotCompleteDiv] =
    useState(false);

  useEffect(() => {
    if (
      (dataOptions?.dataMode !== DashboardDataModeCLISnapshot &&
        dataOptions?.dataMode !== DashboardDataModeCloudSnapshot) ||
      state.state !== "complete"
    ) {
      return;
    }
    setRenderSnapshotCompleteDiv(true);
  }, [dataOptions?.dataMode, state.state]);

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
        render: {
          headless: renderOptions?.headless,
          snapshotCompleteDiv: renderSnapshotCompleteDiv,
        },
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

export { DashboardContext, DashboardProvider, useDashboard };
