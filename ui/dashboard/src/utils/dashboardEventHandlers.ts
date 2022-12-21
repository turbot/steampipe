import { addPanelLog, panelLogTitle } from "./state";
import {
  DashboardActions,
  DashboardExecutionCompleteEvent,
  DashboardExecutionEventWithSchema,
  PanelDefinition,
} from "../types";
import { LATEST_EXECUTION_SCHEMA_VERSION } from "../constants/versions";

const migrateSnapshotDataToExecutionCompleteEvent = (snapshot) => {
  switch (snapshot.schema_version) {
    case "20220614":
    case "20220929":
      const {
        layout,
        panels,
        inputs,
        variables,
        search_path,
        start_time,
        end_time,
      } = snapshot;
      return {
        action: DashboardActions.EXECUTION_COMPLETE,
        schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
        snapshot: {
          schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
          layout,
          panels,
          inputs,
          variables,
          search_path,
          start_time,
          end_time,
        },
      };
    default:
      throw new Error(
        `Unsupported dashboard event schema ${snapshot.schema_version}`
      );
  }
};

const migrateDashboardExecutionCompleteSchema = (
  event: DashboardExecutionEventWithSchema
): DashboardExecutionCompleteEvent => {
  switch (event.schema_version) {
    case "20220614":
      const {
        action,
        execution_id,
        layout,
        panels,
        inputs,
        variables,
        search_path,
        start_time,
        end_time,
      } = event;
      return {
        action,
        schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
        execution_id,
        snapshot: {
          schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
          layout,
          panels,
          inputs,
          variables,
          search_path,
          start_time,
          end_time,
        },
      };
    case LATEST_EXECUTION_SCHEMA_VERSION:
      // Nothing to do here as this event is already in the latest supported schema
      return event as DashboardExecutionCompleteEvent;
    default:
      throw new Error(
        `Unsupported dashboard event schema ${event.schema_version}`
      );
  }
};

const updatePanelsMapWithControlEvent = (panelsMap, action) => {
  return {
    ...panelsMap,
    [action.control.name]: action.control,
  };
};

const controlsUpdatedEventHandler = (action, state) => {
  // If the dashboard has already completed,
  // no need to process these events
  if (
    state.state === "complete" ||
    !action.controls ||
    action.controls.length === 0
  ) {
    return state;
  }

  let panelsMap = state.panelsMap;
  for (const event of action.controls) {
    // We're not expecting execution events for this ID
    if (event.execution_id !== state.execution_id) {
      continue;
    }

    const updatedPanelsMap = updatePanelsMapWithControlEvent(panelsMap, event);

    if (!updatedPanelsMap) {
      continue;
    }

    panelsMap = updatedPanelsMap;
  }
  return {
    ...state,
    panelsMap,
    progress: calculateProgress(panelsMap),
  };
};

const leafNodesUpdatedEventHandler = (action, state) => {
  // If there's nothing to process, no need to mutate state
  if (!action || !action.nodes || action.nodes.length === 0) {
    return state;
  }

  let panelsLog = state.panelsLog;
  let panelsMap = state.panelsMap;
  for (const event of action.nodes) {
    // We're not expecting execution events for this ID
    if (event.execution_id !== state.execution_id) {
      continue;
    }

    const { dashboard_node, timestamp } = event;

    if (!dashboard_node || !dashboard_node.status || !timestamp) {
      continue;
    }

    const { error, status } = dashboard_node;

    panelsLog = addPanelLog(panelsLog, dashboard_node.name, {
      error,
      status,
      timestamp,
      title: panelLogTitle(dashboard_node),
    });

    panelsMap[dashboard_node.name] = dashboard_node;
  }

  const newState = {
    ...state,
    panelsLog,
    progress: calculateProgress(panelsMap),
  };

  if (state.state !== "complete") {
    newState.panelsMap = { ...panelsMap };
  }

  return newState;
};

const calculateProgress = (panelsMap) => {
  const panels: PanelDefinition[] = Object.values(panelsMap || {});
  let dataPanels = 0;
  let completeDataPanels = 0;
  for (const panel of panels) {
    const isControl = panel.panel_type === "control";
    const isPendingDataPanel =
      panel.panel_type !== "container" && panel.panel_type !== "dashboard";
    if (isControl || isPendingDataPanel) {
      dataPanels += 1;
    }
    if (
      (isControl || isPendingDataPanel) &&
      (panel.status === "complete" || panel.status === "error")
    ) {
      completeDataPanels += 1;
    }
  }
  if (dataPanels === 0) {
    return 100;
  }
  return Math.min(Math.ceil((completeDataPanels / dataPanels) * 100), 100);
};

export {
  controlsUpdatedEventHandler,
  leafNodesUpdatedEventHandler,
  migrateDashboardExecutionCompleteSchema,
  migrateSnapshotDataToExecutionCompleteEvent,
};
