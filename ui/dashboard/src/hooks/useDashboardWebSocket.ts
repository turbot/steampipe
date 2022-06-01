import {
  DashboardActionType,
  DashboardDataMode,
  ElementType,
  IActions,
} from "./useDashboard";
import { useCallback, useEffect, useRef } from "react";

export interface EventHooks {
  [type: DashboardActionType]: (event: any) => Promise<void>;
}

interface ReceivedSocketMessagePayload {
  action: string;
  [key: string]: any;
}

interface ReceivedSocketMessage {
  data: string;
}

interface IWebSocket {
  ready: boolean;
  send: (message: SocketMessage) => void;
}

export const SocketActions: IActions = {
  CLEAR_DASHBOARD: "clear_dashboard",
  SELECT_DASHBOARD: "select_dashboard",
  INPUT_CHANGED: "input_changed",
};

const socketMessages = Object.values(SocketActions);

type SocketMessageType = ElementType<typeof socketMessages>;

interface SocketMessagePayloadInputValues {
  [key: string]: any;
}

interface SocketMessagePayloadDashboard {
  full_name: string;
}

interface SocketMessagePayload {
  dashboard: SocketMessagePayloadDashboard;
  input_values: SocketMessagePayloadInputValues;
  changed_input?: string;
}

export interface SocketMessage {
  action: SocketMessageType;
  payload?: SocketMessagePayload;
}

const getSocketServerUrl = () => {
  // In this scenario the browser will be at http://localhost:3000,
  // therefore we have no idea what host + port the dashboard server
  // is on, so assume it's the default
  if (process.env.NODE_ENV === "development") {
    return "ws://localhost:9194/ws";
  }
  // Otherwise, it's a production build, so use the URL details
  const url = new URL(window.location.toString());
  return `${url.protocol === "https:" ? "wss" : "ws"}://${url.host}/ws`;
};

const createSocket = async (socketFactory): Promise<WebSocket> => {
  if (socketFactory) {
    return await socketFactory();
  }
  return new WebSocket(getSocketServerUrl());
};

const useDashboardWebSocket = (
  dispatch,
  socketFactory,
  eventHooks: EventHooks
): IWebSocket => {
  const webSocket = useRef<WebSocket | null>(null);

  const onSocketError = (evt: any) => {
    console.error(evt);
  };

  const onSocketMessage = (message: ReceivedSocketMessage) => {
    const parsed: ReceivedSocketMessagePayload = JSON.parse(message.data);
    dispatch({ type: parsed.action, ...parsed });
    const hookHandler = eventHooks[parsed.action];
    if (hookHandler) {
      hookHandler(parsed);
    }
  };

  useEffect(() => {
    const doConnect = async () => {
      let keepAliveTimerId: NodeJS.Timeout;
      webSocket.current = await createSocket(socketFactory);
      webSocket.current.onerror = onSocketError;
      webSocket.current.onopen = () => {
        const keepAlive = async () => {
          if (!webSocket.current) {
            return;
          }

          const timeout = 30000;
          if (webSocket.current.readyState === webSocket.current.CLOSED) {
            webSocket.current = await createSocket(socketFactory);
            webSocket.current.onerror = onSocketError;
          }
          if (webSocket.current.readyState === webSocket.current.OPEN) {
            webSocket.current.send(JSON.stringify({ action: "keep_alive" }));
          }
          keepAliveTimerId = setTimeout(keepAlive, timeout);
        };

        if (
          !webSocket.current ||
          webSocket.current.readyState !== webSocket.current.OPEN
        ) {
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
        clearTimeout(keepAliveTimerId);
        webSocket.current && webSocket.current.close();
      };
    };
    doConnect();
  }, []);

  useEffect(() => {
    if (!webSocket.current) {
      return;
    }
    webSocket.current.onmessage = onSocketMessage;
  }, [eventHooks, webSocket.current]);

  const send = useCallback((message) => {
    // TODO log this?
    if (!webSocket.current) {
      return;
    }

    webSocket.current.send(JSON.stringify(message));
  }, []);

  return {
    ready: webSocket.current
      ? webSocket.current.readyState === webSocket.current.OPEN
      : false,
    send,
  };
};

export default useDashboardWebSocket;
