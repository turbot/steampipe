import usePanelControls from "./usePanelControls";
import { createContext, ReactNode, useContext, useState } from "react";
import { IPanelControl } from "../components/dashboards/layout/Panel/PanelControls";
import { PanelDefinition } from "../types";

interface IPanelContext {
  definition: PanelDefinition;
  panelControls: IPanelControl[];
  panelInformation: ReactNode | null;
  showPanelControls: boolean;
  showPanelInformation: boolean;
  setPanelInformation: (information: ReactNode) => void;
  setShowPanelControls: (show: boolean) => void;
  setShowPanelInformation: (show: boolean) => void;
}

const PanelContext = createContext<IPanelContext | null>(null);

const PanelProvider = ({ children, definition, showControls }) => {
  const [showPanelControls, setShowPanelControls] = useState(false);
  const [showPanelInformation, setShowPanelInformation] = useState(false);
  const [panelInformation, setPanelInformation] = useState<ReactNode | null>(
    null
  );
  const { panelControls } = usePanelControls(definition, showControls);
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
