import { DashboardActions, ReceivedSocketMessagePayload } from "../types";
import { useCallback, useEffect, useRef } from "react";

const useDashboardWebSocketEventHandler = (dispatch, eventHooks) => {
  const controlEventBuffer = useRef<ReceivedSocketMessagePayload[]>([]);
  const leafNodesUpdatedEventBuffer = useRef<ReceivedSocketMessagePayload[]>(
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
        case DashboardActions.LEAF_NODE_UPDATED:
          leafNodesUpdatedEventBuffer.current.push(event);
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
        leafNodesUpdatedEventBuffer.current.length === 0
      ) {
        return;
      }
      const controlEventsToProcess = [...controlEventBuffer.current];
      const leafNodeCompleteEventsToProcess = [
        ...leafNodesUpdatedEventBuffer.current,
      ];
      controlEventBuffer.current = [];
      leafNodesUpdatedEventBuffer.current = [];
      if (controlEventsToProcess.length > 0) {
        dispatch({
          type: DashboardActions.CONTROLS_UPDATED,
          controls: controlEventsToProcess,
        });
      }
      if (leafNodeCompleteEventsToProcess.length > 0) {
        dispatch({
          type: DashboardActions.LEAF_NODES_UPDATED,
          nodes: leafNodeCompleteEventsToProcess,
        });
      }
    }, 500);
    return () => clearInterval(interval);
  }, [dispatch]);

  return { eventHandler };
};

export default useDashboardWebSocketEventHandler;
