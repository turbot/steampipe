import { createContext, useContext, useEffect, useState } from "react";
import { FoldedNode } from "../../common";
import { KeyValueStringPairs } from "../../common/types";
import { noop } from "../../../../utils/func";
import { useDashboard } from "../../../../hooks/useDashboard";
import { v4 as uuid } from "uuid";

interface IGraphContext {
  expandNode: (foldedNodes: FoldedNode[], category: string) => void;
  expandedNodes: KeyValueStringPairs;
  layoutId: string;
  recalcLayout: () => void;
}

const GraphContext = createContext<IGraphContext>({
  expandNode: noop,
  expandedNodes: {},
  layoutId: "",
  recalcLayout: noop,
});

const GraphProvider = ({ children }) => {
  const {
    themeContext: { theme },
  } = useDashboard();
  const [layoutId, setLayoutId] = useState(uuid());
  const [expandedNodes, setExpandedNodes] = useState<KeyValueStringPairs>({});

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setLayoutId(uuid()), [theme.name]);

  const recalcLayout = () => {
    setExpandedNodes({});
    setLayoutId(uuid());
  };

  const expandNode = (foldedNodes: FoldedNode[] = [], category: string) => {
    setExpandedNodes((current) => {
      const newExpandedNodes = { ...current };
      for (const foldedNode of foldedNodes) {
        newExpandedNodes[foldedNode.id] = category;
      }
      return newExpandedNodes;
    });
    setLayoutId(uuid());
  };

  return (
    <GraphContext.Provider
      value={{ expandNode, expandedNodes, layoutId, recalcLayout }}
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
