import useDashboardDataMode from "./useDashboardDataMode";
import useDashboardInputs from "./useDashboardInputs";
import useDashboardPageTitle from "./useDashboardPageTitle";
import useDashboardSearch from "./useDashboardSearch";
import useDashboardSnapshot from "./useDashboardSnapshot";
import useDashboardState from "./useDashboardState";
import useDashboardWebSocket from "./useDashboardWebSocket";
import {
  AvailableDashboard,
  AvailableDashboardsDictionary,
  ComponentsMap,
  DashboardAction,
  DashboardActions,
  DashboardDataMode,
  DashboardDefinition,
  DashboardInputs,
  DashboardMetadata,
  DashboardProviderProps,
  DashboardRunState,
  DashboardSearch,
  DashboardTags,
  IBreakpointContext,
  IThemeContext,
  PanelsMap,
} from "../../types/dashboard";
import { buildComponentsMap } from "../../components";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { DashboardSnapshot } from "../../types/snapshot";
import { GlobalHotKeys } from "react-hotkeys";
import { noop } from "../../utils/func";
import { PanelDefinition } from "../../types/panel";
import { ReceivedSocketMessagePayload } from "../../types/webSocket";
import { useSearchParams } from "react-router-dom";

interface IDashboardContext {
  availableDashboardsLoaded: boolean;
  breakpointContext: IBreakpointContext;
  closePanelDetail: () => void;
  components: ComponentsMap;
  dashboardTags: DashboardTags;
  dashboard: DashboardDefinition | null;
  dashboards: AvailableDashboard[];
  dashboardsMap: AvailableDashboardsDictionary;
  dataMode: DashboardDataMode;
  dispatch(action: DashboardAction): void;
  error: any;
  inputs: DashboardInputs;
  metadata: DashboardMetadata | null;
  panelsMap: PanelsMap;
  progress: number;
  search: DashboardSearch;
  state: DashboardRunState;
  selectedDashboard: AvailableDashboard | null;
  selectedPanel: PanelDefinition | null;
  selectedSnapshot: DashboardSnapshot | null;
  snapshotId: string | null;
  themeContext: IThemeContext;
}

const DashboardContext = createContext<IDashboardContext | null>(null);

const DashboardProviderNew = ({
  analyticsContext,
  breakpointContext,
  children,
  componentOverrides = {},
  eventHooks = {},
  featureFlags = [],
  socketUrlFactory,
  stateDefaults = {},
  themeContext,
}: DashboardProviderProps) => {
  const components = buildComponentsMap(componentOverrides);
  const [searchParams, setSearchParams] = useSearchParams();
  const { dataMode } = useDashboardDataMode(searchParams);
  const { inputs } = useDashboardInputs(searchParams);
  const { search } = useDashboardSearch(searchParams, stateDefaults.search);
  const {
    ready: socketReady,
    lastMessage,
    send: sendSocketMessage,
  } = useDashboardWebSocket(dataMode, socketUrlFactory);
  const { state: dashboardState, dispatch: dispatchDashboardActionInner } =
    useDashboardState(
      inputs,
      dataMode,
      analyticsContext,
      socketReady,
      sendSocketMessage,
      searchParams,
      setSearchParams
    );
  const { id: snapshotId } = useDashboardSnapshot(
    searchParams,
    dashboardState,
    dataMode,
    featureFlags
  );
  const dispatchDashboardAction = useCallback(
    (action) => {
      console.log(action.type, action);
      switch (action.type) {
        case DashboardActions.INPUT_VALUES_CLEARED:
          // We're not expecting execution events for this ID
          if (action.execution_id !== dashboardState.execution_id) {
            return;
          }
          for (const input of action.cleared_inputs || []) {
            searchParams.delete(input);
          }
          setSearchParams(searchParams, { replace: true });
          break;
        default:
          dispatchDashboardActionInner(action);
      }
    },
    [searchParams, setSearchParams]
  );

  useDashboardPageTitle(dashboardState.selectedDashboard);

  useEffect(() => {
    if (!lastMessage) {
      return;
    }
    dispatchDashboardAction({
      type: (lastMessage as ReceivedSocketMessagePayload).action,
      ...lastMessage,
    });
    const hookHandler =
      eventHooks[(lastMessage as ReceivedSocketMessagePayload).action];
    if (hookHandler) {
      hookHandler(lastMessage);
    }
  }, [dispatchDashboardAction, eventHooks, lastMessage]);

  const [hotKeysHandlers, setHotKeysHandlers] = useState({
    CLOSE_PANEL_DETAIL: noop,
  });

  const hotKeysMap = {
    CLOSE_PANEL_DETAIL: ["esc"],
  };

  const closePanelDetail = useCallback(() => {
    dispatchDashboardAction({
      type: DashboardActions.SELECT_PANEL,
      panel: null,
    });
  }, [dispatchDashboardAction]);

  useEffect(() => {
    setHotKeysHandlers({
      CLOSE_PANEL_DETAIL: closePanelDetail,
    });
  }, [closePanelDetail]);

  const context = {
    analyticsContext,
    availableDashboardsLoaded: dashboardState.availableDashboardsLoaded,
    breakpointContext,
    closePanelDetail,
    components,
    dashboard: dashboardState.dashboard,
    dashboardTags: dashboardState.dashboardTags,
    dashboards: dashboardState.dashboards,
    dashboardsMap: dashboardState.dashboardsMap,
    dataMode,
    dispatch: dispatchDashboardAction,
    error: dashboardState.error,
    inputs,
    metadata: dashboardState.metadata,
    panelsMap: dashboardState.panelsMap,
    progress: dashboardState.progress,
    search,
    selectedDashboard: dashboardState.selectedDashboard,
    selectedPanel: dashboardState.selectedPanel,
    selectedSnapshot: dashboardState.selectedSnapshot,
    snapshotId,
    state: dashboardState.state,
    themeContext,
  };

  console.log(context);

  return (
    <DashboardContext.Provider value={context}>
      <GlobalHotKeys
        allowChanges
        keyMap={hotKeysMap}
        handlers={hotKeysHandlers}
      />
      {children}
    </DashboardContext.Provider>
  );
};

const useDashboardNew = () => {
  const context = useContext(DashboardContext);
  if (context === undefined) {
    throw new Error("useDashboard must be used within a DashboardContext");
  }
  return context as IDashboardContext;
};

export { DashboardProviderNew, useDashboardNew };
