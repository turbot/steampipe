import difference from "lodash/difference";
import useDeepCompareEffect from "use-deep-compare-effect";
import usePrevious from "../../../../hooks/usePrevious";
import useTemplateRender from "../../../../hooks/useTemplateRender";
import { createContext, useContext, useEffect, useState } from "react";
import { Edge, Node, useReactFlow } from "reactflow";
import { FoldedNode, RowRenderResult } from "../../common/types";
import { noop } from "../../../../utils/func";
import { useDashboard } from "../../../../hooks/useDashboard";
import { v4 as uuid } from "uuid";

export type ExpandedNodeInfo = {
  category: string;
  foldedNodes: FoldedNode[];
};

export type ExpandedNodes = {
  [nodeId: string]: ExpandedNodeInfo;
};

interface IGraphContext {
  collapseNodes: (foldedNodes: FoldedNode[]) => void;
  expandNode: (foldedNodes: FoldedNode[], category: string) => void;
  expandedNodes: ExpandedNodes;
  layoutId: string;
  recalcLayout: () => void;
  renderResults: RowRenderResult;
  setGraphEdges: (edges: Edge[]) => void;
  setGraphNodes: (nodes: Node[]) => void;
}

const GraphContext = createContext<IGraphContext>({
  collapseNodes: noop,
  expandNode: noop,
  expandedNodes: {},
  layoutId: "",
  recalcLayout: noop,
  renderResults: {},
  setGraphEdges: noop,
  setGraphNodes: noop,
});

type PreviousNodesAndEdges = {
  nodes: Node[];
  edges: Edge[];
};

type CategoryNodeMap = {
  [category: string]: Node[];
};

const GraphProvider = ({ children }) => {
  const {
    themeContext: { theme },
  } = useDashboard();
  const { fitView } = useReactFlow();
  const { ready: templateRenderReady, renderTemplates } = useTemplateRender();
  const [layoutId, setLayoutId] = useState(uuid());
  const [graphEdges, setGraphEdges] = useState<Edge[]>([]);
  const [graphNodes, setGraphNodes] = useState<Node[]>([]);
  const [expandedNodes, setExpandedNodes] = useState<ExpandedNodes>({});
  const [renderResults, setRenderResults] = useState<RowRenderResult>({});

  const previousNodesAndEdges = usePrevious<PreviousNodesAndEdges>({
    nodes: graphNodes,
    edges: graphEdges,
  });

  useDeepCompareEffect(() => {
    if (!templateRenderReady) {
      return;
    }

    const doRender = async () => {
      const nodesWithHrefs = graphNodes.filter(
        (n) => n.data && !n.data.isFolded && !!n.data.href
      );
      const nodesByCategory: CategoryNodeMap = {};
      for (const node of nodesWithHrefs) {
        const category = node?.data?.category?.name || null;
        if (!category) {
          // What to do? We have no category for this node
          continue;
        }
        nodesByCategory[category] = nodesByCategory[category] || [];
        nodesByCategory[category].push(node);
      }

      const renderResults: RowRenderResult = {};

      for (const [category, nodes] of Object.entries(nodesByCategory)) {
        const hrefTemplate = nodes[0].data.href;
        const results = await renderTemplates(
          { [category]: hrefTemplate },
          nodes.map((n) => n.data.row_data || {})
        );
        for (let nodeIdx = 0; nodeIdx < nodes.length; nodeIdx++) {
          const node = nodes[nodeIdx];
          if (!node.id) {
            continue;
          }
          renderResults[node.id] = results[nodeIdx][category];
        }
      }
      setRenderResults(renderResults);
    };

    doRender();
  }, [graphNodes, renderTemplates, templateRenderReady]);

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
    const differentNodeIdsNewToOldAllFoldNodes =
      differentNodeIdsNewToOld.length > 0 &&
      differentNodeIdsNewToOld.every((n) => n.startsWith("fold-node."));
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
      !differentNodeIdsNewToOldAllFoldNodes &&
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

  const collapseNodes = (foldedNodes: FoldedNode[] = []) => {
    setExpandedNodes((current) => {
      const newExpandedNodes = { ...current };
      for (const foldedNode of foldedNodes) {
        delete newExpandedNodes[foldedNode.id];
      }
      return newExpandedNodes;
    });
    setLayoutId(uuid());
  };

  const expandNode = (foldedNodes: FoldedNode[] = [], category: string) => {
    setExpandedNodes((current) => {
      const newExpandedNodes = { ...current };
      for (const foldedNode of foldedNodes) {
        newExpandedNodes[foldedNode.id] = {
          category,
          foldedNodes,
        };
      }
      return newExpandedNodes;
    });
    setLayoutId(uuid());
  };

  return (
    <GraphContext.Provider
      value={{
        collapseNodes,
        expandNode,
        expandedNodes,
        layoutId,
        recalcLayout,
        renderResults,
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
