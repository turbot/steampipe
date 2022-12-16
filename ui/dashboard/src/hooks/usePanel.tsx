import usePanelControls from "./usePanelControls";
import { createContext, ReactNode, useContext, useMemo, useState } from "react";
import { IPanelControl } from "../components/dashboards/layout/Panel/PanelControls";
import { PanelDefinition, PanelWithsByStatus } from "../types";
import { useDashboard } from "./useDashboard";

interface IPanelContext {
  definition: PanelDefinition;
  panelControls: IPanelControl[];
  panelInformation: ReactNode | null;
  showPanelControls: boolean;
  showPanelInformation: boolean;
  setPanelInformation: (information: ReactNode) => void;
  setShowPanelControls: (show: boolean) => void;
  setShowPanelInformation: (show: boolean) => void;
  withStatuses: PanelWithsByStatus;
}

const PanelContext = createContext<IPanelContext | null>(null);

const PanelProvider = ({ children, definition, showControls }) => {
  const { panelsMap } = useDashboard();
  const [showPanelControls, setShowPanelControls] = useState(false);
  const [showPanelInformation, setShowPanelInformation] = useState(false);
  const [panelInformation, setPanelInformation] = useState<ReactNode | null>(
    null
  );
  const { panelControls } = usePanelControls(definition, showControls);
  const withStatuses = useMemo<PanelWithsByStatus>(() => {
    if (!definition || !definition.withs || definition.withs.length === 0) {
      return {} as PanelWithsByStatus;
    }
    const withsByStatus: PanelWithsByStatus = {};
    for (const withName of definition.withs) {
      const withPanel = panelsMap[withName];
      if (!withPanel || !withPanel.status) {
        continue;
      }
      const statuses = withsByStatus[withPanel.status] || [];
      statuses[withPanel.status].push(withPanel);
      withsByStatus[withPanel.status] = statuses;
    }
    return withsByStatus;
  }, [definition, panelsMap]);

  return (
    <PanelContext.Provider
      value={{
        definition,
        panelControls,
        panelInformation,
        showPanelControls,
        showPanelInformation,
        setPanelInformation,
        setShowPanelControls,
        setShowPanelInformation,
        withStatuses,
      }}
    >
      {children}
    </PanelContext.Provider>
  );
};

const usePanel = () => {
  const context = useContext(PanelContext);
  if (context === undefined) {
    throw new Error("usePanel must be used within a PanelContext");
  }
  return context as IPanelContext;
};

export { PanelContext, PanelProvider, usePanel };
