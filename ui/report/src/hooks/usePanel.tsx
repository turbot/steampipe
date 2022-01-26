import { createContext, useContext } from "react";
import { PanelDefinition } from "./useReport";

interface PanelContext {
  definition: PanelDefinition;
  dimensions: DOMRect;
  showExpand: boolean;
}

const PanelContext = createContext<PanelContext | null>(null);

const PanelProvider = ({ children, definition, dimensions, showExpand }) => {
  return (
    <PanelContext.Provider value={{ definition, dimensions, showExpand }}>
      {children}
    </PanelContext.Provider>
  );
};

const usePanel = () => {
  const context = useContext(PanelContext);
  if (context === undefined) {
    throw new Error("usePanel must be used within a PanelContext");
  }
  return context as PanelContext;
};

export { PanelContext, PanelProvider, usePanel };
