import { PanelDefinition } from "../types";

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

const leafNodesCompleteEventHandler = (action, state) => {
  // If the dashboard has already completed,
  // no need to process these events
  if (
    state.state === "complete" ||
    !action.nodes ||
    action.nodes.length === 0
  ) {
    return state;
  }

  let panelsMap = state.panelsMap;
  for (const event of action.nodes) {
    // We're not expecting execution events for this ID
    if (event.execution_id !== state.execution_id) {
      continue;
    }

    const { dashboard_node } = event;

    panelsMap = {
      ...panelsMap,
      [dashboard_node.name]: dashboard_node,
    };
  }
  return {
    ...state,
    panelsMap,
    progress: calculateProgress(panelsMap),
  };
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

export { controlsUpdatedEventHandler, leafNodesCompleteEventHandler };
