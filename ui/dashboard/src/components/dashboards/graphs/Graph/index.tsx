import AssetNode from "./AssetNode";
import dagre from "dagre";
import ErrorPanel from "../../Error";
import FloatingEdge from "./FloatingEdge";
import ReactFlow, {
  ControlButton,
  Controls,
  Edge,
  MarkerType,
  Node,
  Position,
  useNodesState,
  useEdgesState,
  useReactFlow,
} from "react-flow-renderer";
import { buildNodesAndEdges, LeafNodeData } from "../../common";
import { getGraphComponent } from "..";
import { GraphProperties, GraphProps } from "../types";
import { Ref, useCallback, useEffect, useMemo, useState } from "react";
import { registerComponent } from "../../index";
import {
  ResetLayoutIcon,
  ZoomIcon,
  ZoomInIcon,
  ZoomOutIcon,
} from "../../../../constants/icons";
import { Theme } from "../../../../hooks/useTheme";
import { TooltipsProvider, useTooltips } from "./Tooltip";
import { useDashboard } from "../../../../hooks/useDashboard";
import { v4 as uuid } from "uuid";

const nodeWidth = 100;
const nodeHeight = 100;

const nodeTypes = {
  asset: AssetNode,
};

const edgeTypes = {
  floating: FloatingEdge,
};

const buildGraphNodesAndEdges = (
  data: LeafNodeData | undefined,
  properties: GraphProperties | undefined,
  namedColors: any
) => {
  if (!data) {
    return {
      nodes: [],
      edges: [],
    };
  }
  const nodesAndEdges = buildNodesAndEdges(
    data,
    properties,
    namedColors,
    false
  );
  const direction = properties?.direction || "TB";
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setGraph({
    rankdir: direction,
    nodesep: direction === "LR" || direction === "RL" ? 15 : 110,
    ranksep: direction === "LR" || direction === "RL" ? 200 : 60,
  });
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  nodesAndEdges.nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });
  nodesAndEdges.edges.forEach((edge) => {
    dagreGraph.setEdge(edge.from_id, edge.to_id);
  });
  dagre.layout(dagreGraph);
  const innerGraph = dagreGraph.graph();
  const nodes: Node[] = [];
  const edges: Edge[] = [];
  for (const node of nodesAndEdges.nodes) {
    const matchingNode = dagreGraph.node(node.id);
    // console.log({ node, dagreNode: matchingNode });
    const matchingCategory = node.category
      ? nodesAndEdges.categories[node.category]
      : null;
    nodes.push({
      type: "asset",
      id: node.id,
      position: { x: matchingNode.x, y: matchingNode.y },
      data: {
        color: matchingCategory ? matchingCategory.color : null,
        href: matchingCategory ? matchingCategory.href : null,
        icon: matchingCategory ? matchingCategory.icon : null,
        label: node.title,
        row_data: node.row_data,
        namedColors,
      },
    });
  }
  for (const edge of nodesAndEdges.edges) {
    const matchingCategory = edge.category
      ? nodesAndEdges.categories[edge.category]
      : null;
    edges.push({
      type: "floating",
      id: edge.id,
      source: edge.from_id,
      target: edge.to_id,
      label: edge.title,
      labelBgPadding: [11, 0],
      markerEnd: {
        color: matchingCategory
          ? matchingCategory.color
          : namedColors.blackScale3,
        width: 20,
        height: 20,
        strokeWidth: 2,
        type: MarkerType.Arrow,
      },
      data: {
        color: matchingCategory ? matchingCategory.color : null,
        row_data: edge.row_data,
        label: edge.title,
        namedColors,
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

  return { nodes, edges, width: innerGraph.width, height: innerGraph.height };
};

const useGraphNodesAndEdges = (
  data: LeafNodeData | undefined,
  properties: GraphProperties | undefined,
  namedColors: {},
  id: number
) => {
  const nodesAndEdges = useMemo(
    () => buildGraphNodesAndEdges(data, properties, namedColors),
    [data, properties, id]
  );
  return {
    nodesAndEdges,
  };
};

const useGraphOptions = (
  props: GraphProps,
  id: number,
  theme: Theme,
  themeWrapperRef: ((instance: null) => void) | Ref<null>
) => {
  // We need to get the theme CSS variable values - these are accessible on the theme root element and below in the tree
  const style = themeWrapperRef
    ? // @ts-ignore
      window.getComputedStyle(themeWrapperRef)
    : null;
  let namedColors;
  if (style) {
    const blackScale3 = style.getPropertyValue("--color-black-scale-3").trim();
    const blackScale4 = style.getPropertyValue("--color-black-scale-4").trim();
    const foreground = style.getPropertyValue("--color-foreground").trim();
    const foregroundLightest = style
      .getPropertyValue("--color-foreground-lightest")
      .trim();
    const alert = style.getPropertyValue("--color-alert").trim();
    const info = style.getPropertyValue("--color-info").trim();
    const ok = style.getPropertyValue("--color-ok").trim();
    namedColors = {
      blackScale3,
      blackScale4,
      foreground,
      foregroundLightest,
      alert,
      info,
      ok,
    };
  } else {
    namedColors = {};
  }

  const { nodesAndEdges } = useGraphNodesAndEdges(
    props.data,
    props.properties,
    namedColors,
    id
  );
  const [nodes, setNodes, onNodesChange] = useNodesState(nodesAndEdges.nodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(nodesAndEdges.edges);
  // const onConnect = useCallback(
  //   (params) => setEdges((eds) => addEdge(params, eds)),
  //   []
  // );

  useEffect(() => {
    // console.log("nodes changes", nodesAndEdges.nodes);
    setNodes(nodesAndEdges.nodes);
  }, [nodesAndEdges.nodes]);

  useEffect(() => {
    // console.log("edges changes", nodesAndEdges.edges);
    setEdges(nodesAndEdges.edges);
  }, [nodesAndEdges.edges]);

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

const ZoomInControl = () => {
  const { zoomIn } = useReactFlow();
  return (
    <ControlButton
      className="bg-dashboard text-foreground border-0"
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
      className="bg-dashboard text-foreground border-0"
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
      className="bg-dashboard text-foreground border-0"
      onClick={() => fitView()}
      title="Fit View"
    >
      <ZoomIcon className="w-5 h-5" />
    </ControlButton>
  );
};

const RecalcLayoutControl = ({ recalc }) => (
  <ControlButton
    className="bg-dashboard text-foreground border-0"
    onClick={() => recalc()}
    title="Reset Layout"
  >
    <ResetLayoutIcon className="w-5 h-5" />
  </ControlButton>
);

const CustomControls = ({ recalcLayout }) => {
  const { fitView } = useReactFlow();
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
      <RecalcLayoutControl
        recalc={() => {
          recalcLayout();
          fitView();
        }}
      />
    </Controls>
  );
};

const Graph = ({ id, props, recalc, theme, themeWrapperRef }) => {
  const graphOptions = useGraphOptions(props, id, theme, themeWrapperRef);
  const { closeTooltips } = useTooltips();
  return (
    <ReactFlow
      nodes={graphOptions.nodes}
      edges={graphOptions.edges}
      onNodesChange={graphOptions.onNodesChange}
      onEdgesChange={graphOptions.onEdgesChange}
      onPaneClick={() => closeTooltips()}
      nodeTypes={nodeTypes}
      // @ts-ignore
      edgeTypes={edgeTypes}
      fitView
      style={{ height: Math.min(600, graphOptions.height) }}
      zoomOnScroll={false}
    >
      <CustomControls recalcLayout={() => recalc()} />
    </ReactFlow>
  );
};

const GraphWrapper = (props: GraphProps) => {
  const [id, setId] = useState(uuid());
  const {
    themeContext: { theme, wrapperRef: themeWrapperRef },
  } = useDashboard();

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setId(uuid()), [theme.name]);

  if (!themeWrapperRef) {
    return null;
  }

  if (!props.data) {
    return null;
  }

  return (
    <TooltipsProvider>
      <Graph
        id={id}
        props={props}
        recalc={() => setId(uuid())}
        theme={theme}
        themeWrapperRef={themeWrapperRef}
      />
    </TooltipsProvider>
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
