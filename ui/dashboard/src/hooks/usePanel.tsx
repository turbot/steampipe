import { createContext, useContext } from "react";
import { PanelDefinition } from "../types";

interface IPanelContext {
  definition: PanelDefinition;
  setZoomIconClassName: (className: string) => void;
  showControls: boolean;
}

const PanelContext = createContext<IPanelContext | null>(null);

const PanelProvider = ({
  children,
  definition,
  showControls,
  setZoomIconClassName,
}) => (
  <PanelContext.Provider
    value={{ definition, showControls, setZoomIconClassName }}
  >
    {children}
  </PanelContext.Provider>
);

const usePanel = () => {
  const context = useContext(PanelContext);
  if (context === undefined) {
    throw new Error("usePanel must be used within a PanelContext");
  }
  return context as IPanelContext;
};

export { PanelContext, PanelProvider, usePanel };
