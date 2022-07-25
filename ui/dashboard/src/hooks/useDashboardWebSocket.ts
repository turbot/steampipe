import isEmpty from "lodash/isEmpty";
import useWebSocket, { ReadyState } from "react-use-websocket";
import {
  DashboardActionType,
  DashboardDataMode,
  IActions,
} from "./useDashboard";
import { useCallback, useEffect, useRef } from "react";

export interface EventHooks {
  [type: DashboardActionType]: (event: any) => Promise<void>;
}

export const SocketActions: IActions = {
  CLEAR_DASHBOARD: "clear_dashboard",
  GET_AVAILABLE_DASHBOARDS: "get_available_dashboards",
  GET_DASHBOARD_METADATA: "get_dashboard_metadata",
  SELECT_DASHBOARD: "select_dashboard",
  INPUT_CHANGED: "input_changed",
};

export type SocketURLFactory = () => Promise<string>;

export interface ReceivedSocketMessagePayload {
  action: string;
  [key: string]: any;
}

const useDashboardWebSocket = (
  dataMode: DashboardDataMode,
  dispatch: (action: any) => void,
  eventHooks: {} | undefined,
  socketUrlFactory?: () => Promise<string>
) => {
  const didUnmount = useRef(false);
  // const [socketUrl, setSocketUrl] = useState<string | null>(
  //   !socketUrlFactory ? getSocketServerUrl() : null
  // );

  const getSocketServerUrl = useCallback(async () => {
    if (socketUrlFactory) {
      return socketUrlFactory();
    }

    // In this scenario the browser will be at http://localhost:3000,
    // therefore we have no idea what host + port the dashboard server
    // is on, so assume it's the default
    if (process.env.NODE_ENV === "development") {
      return "ws://localhost:9194/ws";
    }
    // Otherwise, it's a production build, so use the URL details
    const url = new URL(window.location.toString());
    return `${url.protocol === "https:" ? "wss" : "ws"}://${url.host}/ws`;
  }, [socketUrlFactory]);

  const { lastJsonMessage, readyState, sendJsonMessage } = useWebSocket(
    getSocketServerUrl,
    {
      shouldReconnect: () => {
        /*
            useWebSocket will handle unmounting for you, but this is an example of a
            case in which you would not want it to automatically reconnect
          */
        return !didUnmount.current;
      },
      reconnectAttempts: 10,
      reconnectInterval: 3000,
    },
    dataMode === "live"
  );

  useEffect(() => {
    if (!lastJsonMessage || isEmpty(lastJsonMessage)) {
      return;
    }
    dispatch({
      type: (lastJsonMessage as ReceivedSocketMessagePayload).action,
      ...lastJsonMessage,
    });
    const hookHandler =
      eventHooks &&
      eventHooks[(lastJsonMessage as ReceivedSocketMessagePayload).action];
    if (hookHandler) {
      hookHandler(lastJsonMessage);
    }
  }, [dispatch, eventHooks, lastJsonMessage]);

  useEffect(() => {
    if (readyState !== ReadyState.OPEN || !sendJsonMessage) {
      return;
    }
    sendJsonMessage({ action: SocketActions.GET_DASHBOARD_METADATA });
    sendJsonMessage({ action: SocketActions.GET_AVAILABLE_DASHBOARDS });
  }, [readyState, sendJsonMessage]);

  useEffect(() => {
    return () => {
      didUnmount.current = true;
    };
  }, []);

  return {
    ready: readyState === ReadyState.OPEN,
    lastMessage: lastJsonMessage,
    send: sendJsonMessage,
  };
};

export default useDashboardWebSocket;
