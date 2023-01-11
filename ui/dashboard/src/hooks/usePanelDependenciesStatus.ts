import { PanelDefinition } from "../types";
import { PanelDependencyStatuses } from "../components/dashboards/common/types";
import { useMemo } from "react";
import { usePanel } from "./usePanel";

const usePanelDependenciesStatus = () => {
  const { dependenciesByStatus, inputPanelsAwaitingValue } = usePanel();

  return useMemo<PanelDependencyStatuses>(() => {
    const initializedPanels: PanelDefinition[] = [];
    const blockedPanels: PanelDefinition[] = [];
    const runningPanels: PanelDefinition[] = [];
    const cancelledPanels: PanelDefinition[] = [];
    const errorPanels: PanelDefinition[] = [];
    const completePanels: PanelDefinition[] = [];
    let total = 0;
    for (const panels of Object.values(dependenciesByStatus)) {
      for (const panel of panels) {
        total += 1;
        if (panel.status === "initialized") {
          initializedPanels.push(panel);
        } else if (panel.status === "blocked") {
          blockedPanels.push(panel);
        } else if (panel.status === "running") {
          runningPanels.push(panel);
        } else if (panel.status === "cancelled") {
          completePanels.push(panel);
        } else if (panel.status === "error") {
          errorPanels.push(panel);
        } else if (panel.status === "complete") {
          completePanels.push(panel);
        }
      }
    }
    const status = {
      initialized: {
        total: initializedPanels.length,
        panels: initializedPanels,
      },
      blocked: {
        total: blockedPanels.length,
        panels: blockedPanels,
      },
      running: {
        total: runningPanels.length,
        panels: runningPanels,
      },
      cancelled: {
        total: cancelledPanels.length,
        panels: cancelledPanels,
      },
      error: {
        total: errorPanels.length,
        panels: errorPanels,
      },
      complete: {
        total: completePanels.length,
        panels: completePanels,
      },
    };
    return {
      total,
      inputsAwaitingValue: inputPanelsAwaitingValue,
      status,
    };
  }, [dependenciesByStatus, inputPanelsAwaitingValue]);
};

export default usePanelDependenciesStatus;
