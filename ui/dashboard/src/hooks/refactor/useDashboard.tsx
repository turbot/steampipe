import useDashboardDataMode from "./useDashboardDataMode";
import useDashboardInputs from "./useDashboardInputs";
import useDashboardSearch from "./useDashboardSearch";
import useDashboardState from "./useDashboardState";
import useDashboardWebSocket from "./useDashboardWebSocket";
import {
  AvailableDashboard,
  AvailableDashboardsDictionary,
  DashboardDataMode,
  DashboardInputs,
  DashboardMetadata,
  DashboardProviderProps,
  DashboardSearch,
} from "../../types/dashboard";
import { createContext, useContext, useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { ReceivedSocketMessagePayload } from "../../types/webSocket";

interface IDashboardContext {
  dashboards: AvailableDashboard[];
  dashboardsMap: AvailableDashboardsDictionary;
  dataMode: DashboardDataMode;
  metadata: DashboardMetadata | null;
  inputs: DashboardInputs;
  search: DashboardSearch;
}

const DashboardContext = createContext<IDashboardContext | null>(null);

const DashboardProviderNew = ({
  children,
  socketUrlFactory,
}: DashboardProviderProps) => {
  const [searchParams, setSearchParams] = useSearchParams();
  const { dataMode } = useDashboardDataMode(searchParams);
  const { inputs } = useDashboardInputs(searchParams);
  const { search } = useDashboardSearch(searchParams);
  const { state: dashboardState, dispatch: dispatchDashboardAction } =
    useDashboardState();
  const {
    ready: socketReady,
    lastMessage,
    send: sendSocketMessage,
  } = useDashboardWebSocket(dataMode, socketUrlFactory);

  useEffect(() => {
    if (!lastMessage) {
      return;
    }
    dispatchDashboardAction({
      type: (lastMessage as ReceivedSocketMessagePayload).action,
      ...lastMessage,
    });
  }, [dispatchDashboardAction, lastMessage]);

  const context = {
    dashboards: dashboardState.dashboards,
    dashboardsMap: dashboardState.dashboardsMap,
    dataMode,
    inputs,
    metadata: dashboardState.metadata,
    search,
  };

  console.log({
    socket: { ready: socketReady },
    sendSocketMessage,
    ...context,
  });

  return (
    <DashboardContext.Provider
      value={{
        dashboards: dashboardState.dashboards,
        dashboardsMap: dashboardState.dashboardsMap,
        dataMode,
        inputs,
        metadata: dashboardState.metadata,
        search,
      }}
    >
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
