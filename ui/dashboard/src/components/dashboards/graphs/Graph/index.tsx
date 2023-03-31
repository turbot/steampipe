import AssetNode from "./AssetNode";
import dagre from "dagre";
import ErrorPanel from "../../Error";
import FloatingEdge from "./FloatingEdge";
import NodeAndEdgePanelInformation from "../../common/NodeAndEdgePanelInformation";
import ReactFlow, {
  ControlButton,
  Controls,
  Edge,
  MarkerType,
  Node,
  Position,
  ReactFlowProvider,
  useNodesState,
  useEdgesState,
  useReactFlow,
} from "reactflow";
import sortBy from "lodash/sortBy";
import useChartThemeColors from "../../../../hooks/useChartThemeColors";
import useNodeAndEdgeData from "../../common/useNodeAndEdgeData";
import {
  buildNodesAndEdges,
  foldNodesAndEdges,
  LeafNodeData,
} from "../../common";
import {
  Category,
  CategoryMap,
  Edge as EdgeType,
  Node as NodeType,
} from "../../common/types";
import {
  DagreRankDir,
  EdgeStatus,
  GraphDirection,
  GraphProperties,
  GraphProps,
  GraphStatuses,
  NodeAndEdgeDataFormat,
  NodeAndEdgeStatus,
  NodeStatus,
  WithStatus,
} from "../types";
import { DashboardRunState } from "../../../../types";
import { ExpandedNodes, GraphProvider, useGraph } from "../common/useGraph";
import { getGraphComponent } from "..";
import { registerComponent } from "../../index";
import {
  ResetLayoutIcon,
  ZoomIcon,
  ZoomInIcon,
  ZoomOutIcon,
} from "../../../../constants/icons";
import { useDashboard } from "../../../../hooks/useDashboard";
import { useEffect, useMemo } from "react";
import { usePanel } from "../../../../hooks/usePanel";
import "reactflow/dist/style.css";

const nodeWidth = 100;
const nodeHeight = 100;

const getGraphDirection = (direction?: GraphDirection | null): DagreRankDir => {
  if (!direction) {
    return "TB";
  }

  switch (direction) {
    case "left_right":
    case "LR":
      return "LR";
    case "top_down":
    case "TB":
      return "TB";
    default:
      return "TB";
  }
};

const getNodeOrEdgeLabel = (
  item: NodeType | EdgeType,
  category: Category | null
) => {
  if (item.isFolded) {
    if (item.title) {
      return item.title;
    } else if (category?.fold?.title) {
      return category.fold.title;
    } else if (category?.title) {
      return category.title;
    } else {
      return category?.name;
    }
  } else {
    if (item.title) {
      return item.title;
    } else if (category?.title) {
      return category.title;
    } else {
      return category?.name;
    }
  }
};

const buildGraphNodesAndEdges = (
  categories: CategoryMap,
  data: LeafNodeData | undefined,
  properties: GraphProperties | undefined,
  themeColors: any,
  expandedNodes: ExpandedNodes,
  status: DashboardRunState
) => {
  if (!data) {
    return {
      nodes: [],
      edges: [],
    };
  }
  let nodesAndEdges = buildNodesAndEdges(
    categories,
    data,
    properties,
    themeColors,
    false
  );

  nodesAndEdges = foldNodesAndEdges(nodesAndEdges, expandedNodes);
  const direction = getGraphDirection(properties?.direction);
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setGraph({
    rankdir: direction,
    nodesep: direction === "LR" ? 15 : 110,
    ranksep: direction === "LR" ? 200 : 60,
  });
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  nodesAndEdges.edges.forEach((edge) => {
    dagreGraph.setEdge(edge.from_id, edge.to_id);
  });
  const finalNodes: NodeType[] = [];
  nodesAndEdges.nodes.forEach((node) => {
    const nodeEdges = dagreGraph.nodeEdges(node.id);
    if (
      status === "complete" ||
      status === "error" ||
      (nodeEdges && nodeEdges.length > 0)
    ) {
      dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
      finalNodes.push(node);
    }
  });
  dagre.layout(dagreGraph);
  const innerGraph = dagreGraph.graph();
  const nodes: Node[] = [];
  const edges: Edge[] = [];
  for (const node of finalNodes) {
    const matchingNode = dagreGraph.node(node.id);
    const matchingCategory = node.category
      ? nodesAndEdges.categories[node.category]
      : null;
    let categoryColor = matchingCategory ? matchingCategory.color : null;
    if (categoryColor === "auto") {
      categoryColor = null;
    }
    nodes.push({
      type: "asset",
      id: node.id,
      dragHandle: ".custom-drag-handle",
      position: { x: matchingNode.x, y: matchingNode.y },
      // height: 70,
      data: {
        category:
          node.category && categories[node.category]
            ? categories[node.category]
            : null,
        color: categoryColor,
        properties: matchingCategory ? matchingCategory.properties : null,
        href: matchingCategory ? matchingCategory.href : null,
        icon: matchingCategory ? matchingCategory.icon : null,
        fold: matchingCategory ? matchingCategory.fold : null,
        isFolded: node.isFolded,
        foldedNodes: node.foldedNodes,
        label: getNodeOrEdgeLabel(node, matchingCategory),
        row_data: node.row_data,
        themeColors,
      },
    });
  }
  for (const edge of nodesAndEdges.edges) {
    // The color rules are:
    // 1) If the target node of the edge specifies a category and that
    //    category specifies a colour of "auto", refer to rule 3).
    // 2) Else if the edge specifies a category and that category specifies a colour,
    //    that colour is used at 100% opacity for both the edge and the label.
    // 3) Else if the target node of the edge specifies a category and that
    //    category specifies a colour, that colour is used at 50% opacity for the
    //    edge and 70% opacity for the label.
    // 4) Else use black scale 4 at 100% opacity for both the edge and the label.

    const matchingCategory = edge.category
      ? nodesAndEdges.categories[edge.category]
      : null;
    let categoryColor = matchingCategory ? matchingCategory.color : null;
    if (categoryColor === "auto") {
      categoryColor = null;
    }

    let targetNodeColor;
    const targetNode = nodesAndEdges.nodeMap[edge.to_id];
    if (targetNode) {
      const targetCategory = nodesAndEdges.categories[targetNode.category];
      if (targetCategory) {
        targetNodeColor = targetCategory.color;
      }
    }
    const color = categoryColor
      ? categoryColor
      : targetNodeColor
      ? targetNodeColor
      : themeColors.blackScale4;
    const labelOpacity = categoryColor ? 1 : targetNodeColor ? 0.7 : 1;
    const lineOpacity = categoryColor ? 1 : targetNodeColor ? 0.7 : 1;
    edges.push({
      type: "floating",
      id: edge.id,
      source: edge.from_id,
      target: edge.to_id,
      label: edge.title,
      labelBgPadding: [11, 0],
      markerEnd: {
        color,
        width: 20,
        height: 20,
        strokeWidth: 1,
        type: MarkerType.Arrow,
      },
      data: {
        category:
          edge.category && categories[edge.category]
            ? categories[edge.category]
            : null,
        color,
        properties: matchingCategory ? matchingCategory.properties : null,
        labelOpacity,
        lineOpacity,
        row_data: edge.row_data,
        label: getNodeOrEdgeLabel(edge, matchingCategory),
        themeColors,
      },
    });
  }

  nodes.forEach((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    node.targetPosition =
      direction === "LR" ? ("left" as Position) : ("top" as Position);
    node.sourcePosition =
      direction === "LR" ? ("right" as Position) : ("bottom" as Position);

    // We are shifting the dagre node position (anchor=center center) to the top left
    // so it matches the React Flow node anchor point (top left).
    node.position = {
      x: nodeWithPosition.x - nodeWidth / 2,
      y: nodeWithPosition.y - nodeHeight / 2,
    };

    return node;
  });

  return {
    nodes,
    edges,
    width: innerGraph.width < 0 ? 0 : innerGraph.width,
    height: innerGraph.height < 0 ? 0 : innerGraph.height,
  };
};

const useGraphOptions = (props: GraphProps) => {
  const { nodesAndEdges } = useGraphNodesAndEdges(
    props.categories,
    props.data,
    props.properties,
    props.status
  );
  const { setGraphEdges, setGraphNodes } = useGraph();
  const [nodes, setNodes, onNodesChange] = useNodesState(nodesAndEdges.nodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(nodesAndEdges.edges);

  useEffect(() => {
    setGraphEdges(edges);
    setGraphNodes(nodes);
  }, [nodes, edges, setGraphNodes, setGraphEdges]);

  useEffect(() => {
    setNodes(nodesAndEdges.nodes);
  }, [nodesAndEdges.nodes, setNodes]);

  useEffect(() => {
    setEdges(nodesAndEdges.edges);
  }, [nodesAndEdges.edges, setEdges]);

  return {
    nodes,
    edges,
    width: nodesAndEdges.width,
    height: nodesAndEdges.height,
    setEdges,
    onNodesChange,
    onEdgesChange,
  };
};

const useGraphNodesAndEdges = (
  categories: CategoryMap,
  data: LeafNodeData | undefined,
  properties: GraphProperties | undefined,
  status: DashboardRunState
) => {
  const { expandedNodes } = useGraph();
  const themeColors = useChartThemeColors();
  const nodesAndEdges = useMemo(
    () =>
      buildGraphNodesAndEdges(
        categories,
        data,
        properties,
        themeColors,
        expandedNodes,
        status
      ),
    [categories, data, expandedNodes, properties, status, themeColors]
  );

  return {
    nodesAndEdges,
  };
};

const ZoomInControl = () => {
  const { zoomIn } = useReactFlow();
  return (
    <ControlButton
      className="bg-dashboard hover:bg-black-scale-2 text-foreground border-0"
      onClick={() => zoomIn()}
      title="Zoom In"
    >
      <ZoomInIcon className="w-5 h-5" />
    </ControlButton>
  );
};

const ZoomOutControl = () => {
  const { zoomOut } = useReactFlow();
  return (
    <ControlButton
      className="bg-dashboard hover:bg-black-scale-2 text-foreground border-0"
      onClick={() => zoomOut()}
      title="Zoom Out"
    >
      <ZoomOutIcon className="w-5 h-5" />
    </ControlButton>
  );
};

const ResetZoomControl = () => {
  const { fitView } = useReactFlow();

  return (
    <ControlButton
      className="bg-dashboard hover:bg-black-scale-2 text-foreground border-0"
      onClick={() => fitView()}
      title="Fit View"
    >
      <ZoomIcon className="w-5 h-5" />
    </ControlButton>
  );
};

const RecalcLayoutControl = () => {
  const { recalcLayout } = useGraph();

  return (
    <ControlButton
      className="bg-dashboard hover:bg-black-scale-2 text-foreground border-0"
      onClick={() => {
        recalcLayout();
      }}
      title="Reset Layout"
    >
      <ResetLayoutIcon className="w-5 h-5" />
    </ControlButton>
  );
};

const CustomControls = () => {
  return (
    <Controls
      className="flex flex-col space-y-px border-0 shadow-0"
      showFitView={false}
      showInteractive={false}
      showZoom={false}
    >
      <ZoomInControl />
      <ZoomOutControl />
      <ResetZoomControl />
      <RecalcLayoutControl />
    </Controls>
  );
};

const useNodeAndEdgePanelInformation = (
  nodeAndEdgeStatus: NodeAndEdgeStatus,
  dataFormat: NodeAndEdgeDataFormat,
  nodes: Node[],
  status: DashboardRunState
) => {
  const { setShowPanelInformation, setPanelInformation } = usePanel();

  const statuses = useMemo<GraphStatuses>(() => {
    const initializedWiths: WithStatus[] = [];
    const initializedNodes: NodeStatus[] = [];
    const initializedEdges: EdgeStatus[] = [];
    const blockedWiths: WithStatus[] = [];
    const blockedNodes: NodeStatus[] = [];
    const blockedEdges: EdgeStatus[] = [];
    const runningWiths: WithStatus[] = [];
    const runningNodes: NodeStatus[] = [];
    const runningEdges: EdgeStatus[] = [];
    const cancelledWiths: WithStatus[] = [];
    const cancelledNodes: NodeStatus[] = [];
    const cancelledEdges: EdgeStatus[] = [];
    const errorWiths: WithStatus[] = [];
    const errorNodes: NodeStatus[] = [];
    const errorEdges: EdgeStatus[] = [];
    const completeWiths: WithStatus[] = [];
    const completeNodes: NodeStatus[] = [];
    const completeEdges: EdgeStatus[] = [];
    if (nodeAndEdgeStatus) {
      for (const withStatus of sortBy(Object.values(nodeAndEdgeStatus.withs), [
        "title",
        "id",
      ])) {
        if (withStatus.state === "initialized") {
          initializedWiths.push(withStatus);
        } else if (withStatus.state === "blocked") {
          blockedWiths.push(withStatus);
        } else if (withStatus.state === "running") {
          runningWiths.push(withStatus);
        } else if (withStatus.state === "cancelled") {
          cancelledWiths.push(withStatus);
        } else if (withStatus.state === "error") {
          errorWiths.push(withStatus);
        } else {
          completeWiths.push(withStatus);
        }
      }

      const sortedNodes = sortBy(nodeAndEdgeStatus.nodes, [
        "title",
        "category.title",
        "category.name",
        "id",
      ]);
      for (let idx = 0; idx < sortedNodes.length; idx++) {
        const node = sortedNodes[idx] as NodeStatus;
        if (node.state === "initialized") {
          initializedNodes.push(node);
        } else if (node.state === "blocked") {
          blockedNodes.push(node);
        } else if (node.state === "running") {
          runningNodes.push(node);
        } else if (node.state === "cancelled") {
          cancelledNodes.push(node);
        } else if (node.state === "error") {
          errorNodes.push(node);
        } else {
          completeNodes.push(node);
        }
      }

      const sortedEdges = sortBy(nodeAndEdgeStatus.edges, [
        "title",
        "category.title",
        "category.name",
        "id",
      ]);
      for (let idx = 0; idx < sortedEdges.length; idx++) {
        const edge = sortedEdges[idx] as EdgeStatus;
        if (edge.state === "initialized") {
          initializedEdges.push(edge);
        } else if (edge.state === "blocked") {
          blockedEdges.push(edge);
        } else if (edge.state === "running") {
          runningEdges.push(edge);
        } else if (edge.state === "cancelled") {
          cancelledEdges.push(edge);
        } else if (edge.state === "error") {
          errorEdges.push(edge);
        } else {
          completeEdges.push(edge);
        }
      }
    }
    return {
      initialized: {
        total:
          initializedWiths.length +
          initializedNodes.length +
          initializedEdges.length,
        withs: initializedWiths,
        nodes: initializedNodes,
        edges: initializedEdges,
      },
      blocked: {
        total: blockedWiths.length + blockedNodes.length + blockedEdges.length,
        withs: blockedWiths,
        nodes: blockedNodes,
        edges: blockedEdges,
      },
      running: {
        total: runningWiths.length + runningNodes.length + runningEdges.length,
        withs: runningWiths,
        nodes: runningNodes,
        edges: runningEdges,
      },
      cancelled: {
        total:
          cancelledWiths.length + cancelledNodes.length + cancelledEdges.length,
        withs: cancelledWiths,
        nodes: cancelledNodes,
        edges: cancelledEdges,
      },
      error: {
        total: errorWiths.length + errorNodes.length + errorEdges.length,
        withs: errorWiths,
        nodes: errorNodes,
        edges: errorEdges,
      },
      complete: {
        total:
          completeWiths.length + completeNodes.length + completeEdges.length,
        withs: completeWiths,
        nodes: completeNodes,
        edges: completeEdges,
      },
    };
  }, [nodeAndEdgeStatus]);

  useEffect(() => {
    if (
      !nodeAndEdgeStatus ||
      dataFormat === "LEGACY" ||
      (statuses.initialized.total === 0 &&
        statuses.blocked.total === 0 &&
        statuses.running.total === 0 &&
        statuses.error.total === 0 &&
        status === "complete" &&
        nodes.length > 0)
    ) {
      setShowPanelInformation(false);
      setPanelInformation(null);
      return;
    }
    // @ts-ignore
    setPanelInformation(() => (
      <NodeAndEdgePanelInformation
        nodes={nodes}
        status={status}
        statuses={statuses}
      />
    ));
    setShowPanelInformation(true);
  }, [
    dataFormat,
    nodeAndEdgeStatus,
    nodes,
    status,
    statuses,
    setPanelInformation,
    setShowPanelInformation,
  ]);
};

const Graph = (props) => {
  const { selectedPanel } = useDashboard();
  const graphOptions = useGraphOptions(props);
  useNodeAndEdgePanelInformation(
    props.nodeAndEdgeStatus,
    props.dataFormat,
    graphOptions.nodes,
    props.status
  );

  const nodeTypes = useMemo(
    () => ({
      asset: AssetNode,
    }),
    []
  );

  const edgeTypes = useMemo(
    () => ({
      floating: FloatingEdge,
    }),
    []
  );

  return (
    <div
      style={{
        height: graphOptions.height,
        maxHeight: selectedPanel ? undefined : 600,
        minHeight: 175,
      }}
    >
      <ReactFlow
        // @ts-ignore
        edgeTypes={edgeTypes}
        edges={graphOptions.edges}
        fitView
        nodes={graphOptions.nodes}
        nodeTypes={nodeTypes}
        onEdgesChange={graphOptions.onEdgesChange}
        onNodesChange={graphOptions.onNodesChange}
        preventScrolling={false}
        proOptions={{
          account: "paid-pro",
          hideAttribution: true,
        }}
        zoomOnScroll={false}
      >
        <CustomControls />
      </ReactFlow>
    </div>
  );
};

const GraphWrapper = (props: GraphProps) => {
  const nodeAndEdgeData = useNodeAndEdgeData(
    props.data,
    props.properties,
    props.status
  );

  if (!nodeAndEdgeData) {
    return null;
  }

  return (
    <ReactFlowProvider>
      <GraphProvider>
        <Graph
          {...props}
          categories={nodeAndEdgeData.categories}
          data={nodeAndEdgeData.data}
          dataFormat={nodeAndEdgeData.dataFormat}
          properties={nodeAndEdgeData.properties}
          nodeAndEdgeStatus={nodeAndEdgeData.status}
        />
      </GraphProvider>
    </ReactFlowProvider>
  );
};

const renderGraph = (definition: GraphProps) => {
  // We default to sankey diagram if not specified
  const { display_type = "graph" } = definition;

  const graph = getGraphComponent(display_type);

  if (!graph) {
    return <ErrorPanel error={`Unknown graph type ${display_type}`} />;
  }

  const Component = graph.component;
  return <Component {...definition} />;
};

const RenderGraph = (props: GraphProps) => {
  return renderGraph(props);
};

registerComponent("graph", RenderGraph);

export default GraphWrapper;
