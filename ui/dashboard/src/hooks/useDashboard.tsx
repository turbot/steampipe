import get from "lodash/get";
import has from "lodash/has";
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
  definition: DashboardDefinition,
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

const updatePanelsMapWithControlEvent = (panelsMap, action) => {
  return {
    ...panelsMap,
    [action.control.name]: action.control,
  };
};

const calculateProgress = (panelsMap) => {
  const panels: PanelDefinition[] = Object.values(panelsMap || {});
  let dataPanels = 0;
  let completeDataPanels = 0;
  for (const panel of panels) {
    const isControl = panel.panel_type === "control";
    const isDataPanel = has(panel, "sql");
    if (isControl || isDataPanel) {
      dataPanels += 1;
    }
    if (
      (isControl &&
        (panel.status === "complete" || panel.status === "error")) ||
      (isDataPanel && has(panel, "data"))
    ) {
      completeDataPanels += 1;
    }
  }
  if (dataPanels === 0) {
    return 100;
  }
  return Math.min(Math.ceil((completeDataPanels / dataPanels) * 100), 100);
};

function reducer(state, action) {
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

      // Migrate from old format
      if (!action.snapshot) {
        const { action: eventAction, dashboard_node, ...rest } = action;
        action.snapshot = {
          ...rest,
        };
      }

      const layout = action.snapshot.layout;
      const panels = action.snapshot.panels;
      const rootLayoutPanel = action.snapshot.layout;
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
    case DashboardActions.SELECT_PANEL:
      return { ...state, selectedPanel: action.panel };
    case DashboardActions.CLEAR_SNAPSHOT:
      return {
        ...state,
        selectedSnapshot: null,
        snapshotId: null,
        dataMode: DashboardDataModeLive,
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
    })
  );
  const dispatch = useCallback((action) => {
    // console.log(action.type, action);
    dispatchInner(action);
  }, []);
  const { dashboard_name } = useParams();
  const { ready: socketReady, send: sendSocketMessage } = useDashboardWebSocket(
    state.dataMode,
    dispatch,
    eventHooks,
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
    if (state.dataMode !== DashboardDataModeLive) {
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
        // @ts-ignore
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
      state.dataMode === DashboardDataModeCLISnapshot ||
      state.dataMode === DashboardDataModeCLISnapshot ||
      (previousSelectedDashboardStates &&
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
      (state.dataMode === DashboardDataModeLive ||
        state.dataMode === DashboardDataModeCLISnapshot) &&
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

export { DashboardActions, DashboardContext, DashboardProvider, useDashboard };
