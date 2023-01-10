import { InputProperties } from "../components/dashboards/inputs/types";
import { PanelDependencyStatuses } from "../components/dashboards/common/types";
import { PanelDefinition } from "../types";
import { useDashboard } from "./useDashboard";
import { useMemo } from "react";
import { usePanel } from "./usePanel";

const usePanelDependenciesStatus = () => {
  const { dependenciesByStatus } = usePanel();
  const { selectedDashboardInputs } = useDashboard();
  return useMemo<PanelDependencyStatuses>(() => {
    const inputPanelsAwaitingValue: PanelDefinition[] = [];
    const initializedPanels: PanelDefinition[] = [];
    const blockedPanels: PanelDefinition[] = [];
    const runningPanels: PanelDefinition[] = [];
    const cancelledPanels: PanelDefinition[] = [];
    const errorPanels: PanelDefinition[] = [];
    const completePanels: PanelDefinition[] = [];
    let total = 0;
    for (const panels of Object.values(dependenciesByStatus)) {
      for (const panel of panels) {
        const isInput = panel.panel_type === "input";
        const inputProperties = isInput
          ? (panel.properties as InputProperties)
          : null;
        const hasInputValue =
          isInput &&
          inputProperties?.unqualified_name &&
          !!selectedDashboardInputs[inputProperties?.unqualified_name];
        total += 1;
        if (isInput && !hasInputValue) {
          inputPanelsAwaitingValue.push(panel);
        }
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
  }, [dependenciesByStatus, selectedDashboardInputs]);
};

export default usePanelDependenciesStatus;
