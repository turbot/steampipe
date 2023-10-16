import { createContext, useContext, useState } from "react";
import { noop } from "../../../../utils/func";
import { CheckGroupingContext } from "../../../../hooks/useCheckGrouping";

type IDashboardControlsContext = {
  context: any;
  setContext: (context: any) => void;
};

const DashboardControlsContext = createContext<IDashboardControlsContext>({
  context: null,
  setContext: noop,
});

const DashboardControlsProvider = ({ children }) => {
  const [context, setContext] = useState(null);

  return (
    <DashboardControlsContext.Provider value={{ context, setContext }}>
      {children}
    </DashboardControlsContext.Provider>
  );
};

const useDashboardControls = () => {
  const context = useContext(DashboardControlsContext);
  if (context === undefined) {
    throw new Error(
      "useDashboardControls must be used within a DashboardControlsContext",
    );
  }
  return context as IDashboardControlsContext;
};

export { DashboardControlsProvider, useDashboardControls };
