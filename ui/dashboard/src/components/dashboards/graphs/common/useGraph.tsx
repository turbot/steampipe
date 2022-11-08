import { createContext, useContext, useEffect, useState } from "react";
import { Edge, Node } from "react-flow-renderer";
import { FoldedNode, KeyValueStringPairs } from "../../common/types";
import { noop } from "../../../../utils/func";
import { useDashboard } from "../../../../hooks/useDashboard";
import { v4 as uuid } from "uuid";

interface IGraphContext {
  expandNode: (foldedNodes: FoldedNode[], category: string) => void;
  expandedNodes: KeyValueStringPairs;
  layoutId: string;
  recalcLayout: () => void;
  setFitView: (fitView: typeof noop) => void;
  setGraphEdges: (edges: Edge[]) => void;
  setGraphNodes: (nodes: Node[]) => void;
}

const GraphContext = createContext<IGraphContext>({
  expandNode: noop,
  expandedNodes: {},
  layoutId: "",
  recalcLayout: noop,
  setFitView: noop,
  setGraphEdges: noop,
  setGraphNodes: noop,
});

const GraphProvider = ({ children }) => {
  const {
    themeContext: { theme },
  } = useDashboard();
  const [layoutId, setLayoutId] = useState(uuid());
  const [fitView, setFitView] = useState<typeof noop>(noop);
  const [graphEdges, setGraphEdges] = useState<Edge[]>([]);
  const [graphNodes, setGraphNodes] = useState<Node[]>([]);
  const [expandedNodes, setExpandedNodes] = useState<KeyValueStringPairs>({});

  // When the edges or nodes change, update the layout
  useEffect(() => {
    if (!fitView || (!graphEdges && !graphNodes)) {
      return;
    }
    fitView();
  }, [fitView, graphEdges, graphNodes]);

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
      value={{
        expandNode,
        expandedNodes,
        layoutId,
        recalcLayout,
        setFitView: (newFitView) => setFitView(() => newFitView),
        setGraphEdges,
        setGraphNodes,
      }}
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
