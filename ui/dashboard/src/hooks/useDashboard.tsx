import get from "lodash/get";
import isEqual from "lodash/isEqual";
import useDashboardState from "./useDashboardState";
import useDashboardWebSocket, { SocketActions } from "./useDashboardWebSocket";
import useDashboardWebSocketEventHandler from "./useDashboardWebSocketEventHandler";
import usePrevious from "./usePrevious";
import {
  DashboardActions,
  DashboardDataModeCLISnapshot,
  DashboardDataModeCloudSnapshot,
  DashboardDataModeLive,
  DashboardDataOptions,
  DashboardRenderOptions,
  IDashboardContext,
  SelectedDashboardStates,
  SocketURLFactory,
} from "../types";
import { buildComponentsMap } from "../components";
import { buildSelectedDashboardInputsFromSearchParams } from "../utils/state";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { GlobalHotKeys } from "react-hotkeys";
import { noop } from "../utils/func";
import {
  useLocation,
  useNavigate,
  useNavigationType,
  useParams,
  useSearchParams,
} from "react-router-dom";

const DashboardContext = createContext<IDashboardContext | null>(null);

type DashboardProviderProps = {
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
};

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
  const [state, dispatch] = useDashboardState({
    dataOptions,
    renderOptions,
    searchParams,
    stateDefaults,
    versionMismatchCheck,
  });
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
