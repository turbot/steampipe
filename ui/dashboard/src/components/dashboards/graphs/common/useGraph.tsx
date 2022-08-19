import { createContext, useContext, useEffect, useState } from "react";
import { KeyValuePairs } from "../../common/types";
import { noop } from "../../../../utils/func";
import { useDashboard } from "../../../../hooks/useDashboard";
import { v4 as uuid } from "uuid";

interface IGraphContext {
  expandCategory: (category: string) => void;
  expandedCategories: KeyValuePairs;
  layoutId: string;
  recalcLayout: () => void;
}

const GraphContext = createContext<IGraphContext>({
  expandCategory: noop,
  expandedCategories: {},
  layoutId: "",
  recalcLayout: noop,
});

const GraphProvider = ({ children }) => {
  const {
    themeContext: { theme },
  } = useDashboard();
  const [layoutId, setLayoutId] = useState(uuid());
  const [expandedCategories, setExpandedCategories] = useState<KeyValuePairs>(
    {}
  );

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setLayoutId(uuid()), [theme.name]);

  const recalcLayout = () => {
    setExpandedCategories({});
    setLayoutId(uuid());
  };

  const expandCategory = (category: string) => {
    setExpandedCategories((current) => ({ ...current, [category]: true }));
  };

  return (
    <GraphContext.Provider
      value={{ expandCategory, expandedCategories, layoutId, recalcLayout }}
    >
      {children}
    </GraphContext.Provider>
  );
};

const useGraph = () => {
  const context = useContext(GraphContext);
  if (context === undefined) {
    throw new Error("useGraph must be used within a GraphContext");
  }
  return context as IGraphContext;
};

export { GraphProvider, useGraph };
