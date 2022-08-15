import AssetNode from "./AssetNode";
import dagre from "dagre";
import ErrorPanel from "../../Error";
import FloatingEdge from "./FloatingEdge";
import ReactFlow, {
  Controls,
  Edge,
  Node,
  Position,
  useNodesState,
  useEdgesState,
  MarkerType,
  addEdge,
} from "react-flow-renderer";
import { buildNodesAndEdges, LeafNodeData } from "../../common";
import { getGraphComponent } from "..";
import { GraphProperties, GraphProps } from "../types";
import { registerComponent } from "../../index";
import { Ref, useCallback, useEffect, useMemo, useState } from "react";
import { Theme } from "../../../../hooks/useTheme";
import { useDashboard } from "../../../../hooks/useDashboard";

const nodeWidth = 70;
const nodeHeight = 70;

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
  const nodesAndEdges = buildNodesAndEdges(data, properties, namedColors);
  const direction = properties?.direction || "TB";
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setGraph({ rankdir: direction });
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
    const matchingCategory = node.category
      ? nodesAndEdges.categories[node.category]
      : null;
    nodes.push({
      type: "asset",
      id: node.id,
      position: { x: matchingNode.x, y: matchingNode.y },
      data: {
        href: matchingCategory ? matchingCategory.href : null,
        icon: matchingCategory ? matchingCategory.icon : null,
        label: node.title,
        row_data: node.row_data,
      },
    });
  }
  for (const edge of nodesAndEdges.edges) {
    edges.push({
      type: "floating",
      id: edge.id,
      source: edge.from_id,
      target: edge.to_id,
      label: edge.title,
      labelBgPadding: [11, 0],
      markerEnd: {
        type: MarkerType.ArrowClosed,
      },
      data: {
        row_data: edge.row_data,
        label: edge.title,
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

const useGraphOptions = (
  props: GraphProps,
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
    const foreground = style.getPropertyValue("--color-foreground");
    const foregroundLightest = style.getPropertyValue(
      "--color-foreground-lightest"
    );
    const alert = style.getPropertyValue("--color-alert");
    const info = style.getPropertyValue("--color-info");
    const ok = style.getPropertyValue("--color-ok");
    namedColors = {
      foreground,
      foregroundLightest,
      alert,
      info,
      ok,
    };
  } else {
    namedColors = {};
  }

  const nodesAndEdges = useMemo(
    () => buildGraphNodesAndEdges(props.data, props.properties, namedColors),
    [props.data, props.properties]
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

const Graph = ({ props, theme, themeWrapperRef }) => {
  const graphOptions = useGraphOptions(props, theme, themeWrapperRef);
  const onConnect = useCallback(
    (params) =>
      graphOptions.setEdges((eds) =>
        addEdge(
          {
            ...params,
            type: "floating",
            markerEnd: { type: MarkerType.Arrow },
          },
          eds
        )
      ),
    [graphOptions.setEdges]
  );
  return (
    <ReactFlow
      nodes={graphOptions.nodes}
      edges={graphOptions.edges}
      onConnect={onConnect}
      onNodesChange={graphOptions.onNodesChange}
      onEdgesChange={graphOptions.onEdgesChange}
      nodeTypes={nodeTypes}
      // @ts-ignore
      edgeTypes={edgeTypes}
      fitView
      style={{ height: Math.min(600, graphOptions.height) }}
      zoomOnScroll={false}
    >
      <Controls />
    </ReactFlow>
  );
};

const GraphWrapper = (props: GraphProps) => {
  const [, setRandomVal] = useState(0);
  const {
    themeContext: { theme, wrapperRef: themeWrapperRef },
  } = useDashboard();

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  if (!themeWrapperRef) {
    return null;
  }

  if (!props.data) {
    return null;
  }

  return (
    <Graph props={props} theme={theme} themeWrapperRef={themeWrapperRef} />
  );

  // return (
  //   <Chart
  //     options={buildGraphOptions(props, theme, wrapperRef)}
  //     type={props.display_type || "graph"}
  //   />
  // );
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
