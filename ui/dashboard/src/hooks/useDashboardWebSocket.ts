import { ElementType, IActions } from "./useDashboard";
import { useCallback, useEffect, useRef } from "react";

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
  return `ws://${url.host}/ws`;
};

const createSocket = (socketFactory): WebSocket => {
  if (socketFactory) {
    return socketFactory();
  }
  return new WebSocket(getSocketServerUrl());
};

const useDashboardWebSocket = (dispatch, socketFactory): IWebSocket => {
  const webSocket = useRef<WebSocket | null>(null);

  const onSocketError = (evt: any) => {
    console.error(evt);
  };

  const onSocketMessage = (message: ReceivedSocketMessage) => {
    const { action, ...rest }: ReceivedSocketMessagePayload = JSON.parse(
      message.data
    );
    dispatch({ type: action, ...rest });
  };

  useEffect(() => {
    let keepAliveTimerId: NodeJS.Timeout;
    webSocket.current = createSocket(socketFactory);
    webSocket.current.onerror = onSocketError;
    webSocket.current.onmessage = onSocketMessage;
    webSocket.current.onopen = () => {
      const keepAlive = () => {
        if (!webSocket.current) {
          return;
        }

        const timeout = 30000;
        if (webSocket.current.readyState === webSocket.current.CLOSED) {
          webSocket.current = createSocket(socketFactory);
          webSocket.current.onerror = onSocketError;
          webSocket.current.onmessage = onSocketMessage;
        }
        if (webSocket.current.readyState === webSocket.current.OPEN) {
          webSocket.current.send(JSON.stringify({ action: "keep_alive" }));
        }
        keepAliveTimerId = setTimeout(keepAlive, timeout);
      };

      if (!webSocket.current) {
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
  }, []);

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
