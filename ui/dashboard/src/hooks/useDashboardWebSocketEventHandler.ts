import { DashboardActions, ReceivedSocketMessagePayload } from "../types";
import { useCallback, useEffect, useRef } from "react";

const useDashboardWebSocketEventHandler = (dispatch, eventHooks) => {
  const controlEventBuffer = useRef<ReceivedSocketMessagePayload[]>([]);
  const leafNodeCompleteEventBuffer = useRef<ReceivedSocketMessagePayload[]>(
    []
  );

  const eventHandler = useCallback(
    (event: ReceivedSocketMessagePayload) => {
      switch (event.action) {
        case DashboardActions.CONTROL_COMPLETE:
        case DashboardActions.CONTROL_ERROR:
          controlEventBuffer.current.push(event);
          break;
        case DashboardActions.LEAF_NODE_COMPLETE:
          leafNodeCompleteEventBuffer.current.push(event);
          break;
        default:
          dispatch({
            type: event.action,
            ...event,
          });
      }

      const hookHandler = eventHooks && eventHooks[event.action];
      if (hookHandler) {
        hookHandler(event);
      }
    },
    [dispatch, eventHooks]
  );

  useEffect(() => {
    const interval = setInterval(() => {
      if (
        controlEventBuffer.current.length === 0 &&
        leafNodeCompleteEventBuffer.current.length === 0
      ) {
        return;
      }
      const controlEventsToProcess = [...controlEventBuffer.current];
      const leafNodeCompleteEventsToProcess = [
        ...leafNodeCompleteEventBuffer.current,
      ];
      controlEventBuffer.current = [];
      leafNodeCompleteEventBuffer.current = [];
      if (controlEventsToProcess.length > 0) {
        dispatch({
          type: DashboardActions.CONTROLS_UPDATED,
          controls: controlEventsToProcess,
        });
      }
      if (leafNodeCompleteEventsToProcess.length > 0) {
        dispatch({
          type: DashboardActions.LEAF_NODES_COMPLETE,
          nodes: leafNodeCompleteEventsToProcess,
        });
      }
    }, 500);
    return () => clearInterval(interval);
  }, [dispatch]);

  return { eventHandler };
};

export default useDashboardWebSocketEventHandler;
