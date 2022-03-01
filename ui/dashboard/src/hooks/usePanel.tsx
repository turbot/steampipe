import { createContext, useContext } from "react";
import { PanelDefinition } from "./useDashboard";

interface IPanelContext {
  definition: PanelDefinition;
  allowExpand: boolean;
  setZoomIconClassName: (className: string) => void;
}

const PanelContext = createContext<IPanelContext | null>(null);

const PanelProvider = ({
  children,
  definition,
  allowExpand,
  setZoomIconClassName,
}) => {
  return (
    <PanelContext.Provider
      value={{ definition, allowExpand, setZoomIconClassName }}
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
