import difference from "lodash/difference";
import usePrevious from "../../../../hooks/usePrevious";
import { createContext, useContext, useEffect, useState } from "react";
import { Edge, Node, useReactFlow } from "reactflow";
import { FoldedNode, KeyValueStringPairs } from "../../common/types";
import { noop } from "../../../../utils/func";
import { useDashboard } from "../../../../hooks/useDashboard";
import { v4 as uuid } from "uuid";

interface IGraphContext {
  expandNode: (foldedNodes: FoldedNode[], category: string) => void;
  expandedNodes: KeyValueStringPairs;
  layoutId: string;
  recalcLayout: () => void;
  setGraphEdges: (edges: Edge[]) => void;
  setGraphNodes: (nodes: Node[]) => void;
}

const GraphContext = createContext<IGraphContext>({
  expandNode: noop,
  expandedNodes: {},
  layoutId: "",
  recalcLayout: noop,
  setGraphEdges: noop,
  setGraphNodes: noop,
});

type PreviousNodesAndEdges = {
  nodes: Node[];
  edges: Edge[];
};

const GraphProvider = ({ children }) => {
  const {
    themeContext: { theme },
  } = useDashboard();
  const { fitView } = useReactFlow();
  const [layoutId, setLayoutId] = useState(uuid());
  const [graphEdges, setGraphEdges] = useState<Edge[]>([]);
  const [graphNodes, setGraphNodes] = useState<Node[]>([]);
  const [expandedNodes, setExpandedNodes] = useState<KeyValueStringPairs>({});

  const previousNodesAndEdges = usePrevious<PreviousNodesAndEdges>({
    nodes: graphNodes,
    edges: graphEdges,
  });

  // When the edges or nodes change, update the layout
  useEffect(() => {
    if (!fitView || (!graphEdges && !graphNodes) || !previousNodesAndEdges) {
      return;
    }
    const previousNodeIds = previousNodesAndEdges.nodes.map((n) => n.id);
    const currentNodeIds = graphNodes.map((n) => n.id);
    const previousEdgeIds = previousNodesAndEdges.edges.map((e) => e.id);
    const currentEdgeIds = graphEdges.map((e) => e.id);
    const expandedNodesKeys = Object.keys(expandedNodes);
    const differentNodeIdsOldToNew = difference(
      previousNodeIds,
      currentNodeIds
    );
    const differentNodeIdsOldToNewAllFoldNodes =
      differentNodeIdsOldToNew.length > 0 &&
      differentNodeIdsOldToNew.every((n) => n.startsWith("fold-node."));
    const differentNodeIdsNewToOld = difference(
      currentNodeIds,
      previousNodeIds
    );
    const differentNodeIdsNewToOldWithoutExpanded = difference(
      differentNodeIdsNewToOld,
      expandedNodesKeys
    );
    const differentEdgeIdsOldToNew = difference(
      previousEdgeIds,
      currentEdgeIds
    );
    const differentEdgeIdsNewToOld = difference(
      currentEdgeIds,
      previousEdgeIds
    );
    if (
      !differentNodeIdsOldToNewAllFoldNodes &&
      (differentNodeIdsOldToNew.length > 0 ||
        differentNodeIdsNewToOldWithoutExpanded.length > 0 ||
        differentEdgeIdsOldToNew.length > 0 ||
        differentEdgeIdsNewToOld.length > 0)
    ) {
      fitView();
    }
  }, [previousNodesAndEdges, expandedNodes, graphEdges, graphNodes, fitView]);

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
