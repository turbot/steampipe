import { addUpdatedPanelLogs } from "./state";
import {
  EXECUTION_SCHEMA_VERSION_20220614,
  EXECUTION_SCHEMA_VERSION_20220929,
  EXECUTION_SCHEMA_VERSION_20221222,
} from "../constants/versions";
import { PanelDefinition, PanelsMap } from "../types";

const migratePanelStatus = (
  panel: PanelDefinition,
  currentSchemaVersion: string
): PanelDefinition => {
  switch (currentSchemaVersion) {
    case "":
    case EXECUTION_SCHEMA_VERSION_20220614:
    case EXECUTION_SCHEMA_VERSION_20220929:
      return {
        ...panel,
        // @ts-ignore
        status: panel.status === "ready" ? "running" : panel.status,
      };
    case EXECUTION_SCHEMA_VERSION_20221222:
      // Nothing to do - already the latest statuses
      return panel;
    default:
      throw new Error(
        `Unsupported dashboard event schema ${currentSchemaVersion}`
      );
  }
};

const migratePanelStatuses = (
  panelsMap: PanelsMap,
  currentSchemaVersion: string
): PanelsMap => {
  const newPanelsMap = {};
  for (const [name, panel] of Object.entries(panelsMap || {})) {
    newPanelsMap[name] = migratePanelStatus(panel, currentSchemaVersion);
  }
  return newPanelsMap;
};

const updatePanelsMapWithControlEvent = (panelsMap, action) => {
  return {
    ...panelsMap,
    [action.control.name]: migratePanelStatus(action.control, ""),
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

const leafNodesUpdatedEventHandler = (action, currentSchemaVersion, state) => {
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

    if (!dashboard_node || !dashboard_node.status) {
      continue;
    }

    panelsLog = addUpdatedPanelLogs(panelsLog, dashboard_node, timestamp);
    panelsMap[dashboard_node.name] = migratePanelStatus(
      dashboard_node,
      currentSchemaVersion
    );
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
      panel.panel_type !== "container" &&
      panel.panel_type !== "dashboard" &&
      panel.panel_type !== "benchmark";
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
  migratePanelStatus,
  migratePanelStatuses,
};
