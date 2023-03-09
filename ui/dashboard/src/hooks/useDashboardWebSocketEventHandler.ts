import { DashboardActions, ReceivedSocketMessagePayload } from "../types";
import { useCallback, useEffect, useRef } from "react";

const useDashboardWebSocketEventHandler = (dispatch, eventHooks) => {
  const controlEventBuffer = useRef<ReceivedSocketMessagePayload[]>([]);
  const leafNodesCompletedEventBuffer = useRef<ReceivedSocketMessagePayload[]>(
    []
  );
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
          leafNodesCompletedEventBuffer.current.push({
            ...event,
            timestamp: new Date().toISOString(),
          });
          break;
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
        leafNodesCompletedEventBuffer.current.length === 0 &&
        leafNodesUpdatedEventBuffer.current.length === 0
      ) {
        return;
      }
      const controlEventsToProcess = [...controlEventBuffer.current];
      const leafNodeCompletedEventsToProcess = [
        ...leafNodesCompletedEventBuffer.current,
      ];
      const leafNodeUpdatedEventsToProcess = [
        ...leafNodesUpdatedEventBuffer.current,
      ];
      controlEventBuffer.current = [];
      leafNodesCompletedEventBuffer.current = [];
      leafNodesUpdatedEventBuffer.current = [];
      if (controlEventsToProcess.length > 0) {
        dispatch({
          type: DashboardActions.CONTROLS_UPDATED,
          controls: controlEventsToProcess,
        });
      }
      if (leafNodeCompletedEventsToProcess.length > 0) {
        dispatch({
          type: DashboardActions.LEAF_NODES_COMPLETE,
          nodes: leafNodeCompletedEventsToProcess,
        });
      }
      if (leafNodeUpdatedEventsToProcess.length > 0) {
        dispatch({
          type: DashboardActions.LEAF_NODES_UPDATED,
          nodes: leafNodeUpdatedEventsToProcess,
        });
      }
    }, 500);
    return () => clearInterval(interval);
  }, [dispatch]);

  return { eventHandler };
};

export default useDashboardWebSocketEventHandler;
