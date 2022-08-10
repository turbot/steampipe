import dagre from "dagre";
import ErrorPanel from "../../Error";
import ReactFlow, {
  Controls,
  Edge,
  Node,
  Position,
  useNodesState,
  useEdgesState,
  MarkerType,
} from "react-flow-renderer";
import merge from "lodash/merge";
import {
  buildGraphDataInputs,
  buildNodesAndEdges,
  LeafNodeData,
  NodesAndEdges,
  toEChartsType,
} from "../../common";
import { formatChartTooltip } from "../../common/chart";
import { GraphProperties, GraphProps, GraphType } from "../types";
import { getGraphComponent } from "..";
import { registerComponent } from "../../index";
import { useDashboard } from "../../../../hooks/useDashboard";
import { Ref, useEffect, useState } from "react";
import { Theme } from "../../../../hooks/useTheme";
import AssetNode from "./AssetNode";
import FloatingEdge from "./FloatingEdge";
import { usePanel } from "../../../../hooks/usePanel";

const getCommonBaseOptions = () => ({
  animation: false,
  tooltip: {
    appendToBody: true,
    borderWidth: 0,
    padding: 0,
    trigger: "item",
    triggerOn: "mousemove",
  },
});

const getCommonBaseOptionsForGraphType = (type: GraphType = "graph") => {
  switch (type) {
    case "graph":
      return {};
    default:
      return {};
  }
};

const getSeriesForGraphType = (
  type: GraphType = "graph",
  data: LeafNodeData | undefined,
  properties: GraphProperties | undefined,
  nodesAndEdges: NodesAndEdges,
  namedColors
) => {
  if (!data) {
    return {};
  }
  const series: any[] = [];
  const seriesLength = 1;
  for (let seriesIndex = 0; seriesIndex < seriesLength; seriesIndex++) {
    switch (type) {
      case "graph": {
        const { data: graphData, links } = buildGraphDataInputs(nodesAndEdges);
        series.push({
          type: toEChartsType(type),
          layout: "force",
          roam: true,
          draggable: true,
          label: {
            show: true,
            color: namedColors.foreground,
            formatter: "{b}",
            position: "bottom",
          },
          labelLayout: {
            hideOverlap: true,
          },
          scaleLimit: {
            min: 0.4,
            max: 4,
          },
          edgeSymbol: ["none", "arrow"],
          emphasis: {
            focus: "adjacency",
            blurScope: "coordinateSystem",
          },
          lineStyle: {
            color: "source",
            curveness: 0,
          },
          data: graphData,
          links,
          tooltip: {
            formatter: (nodeData) => formatChartTooltip(nodeData),
          },
        });
        break;
      }
    }
  }

  return { series };
};

const getOptionOverridesForGraphType = (
  type: GraphType = "graph",
  properties: GraphProperties | undefined
) => {
  if (!properties) {
    return {};
  }

  return {};
};

const buildGraphOptions = (props: GraphProps, theme, themeWrapperRef) => {
  // We need to get the theme CSS variable values - these are accessible on the theme root element and below in the tree
  // @ts-ignore
  const style = window.getComputedStyle(themeWrapperRef);
  const foreground = style.getPropertyValue("--color-foreground");
  const foregroundLightest = style.getPropertyValue(
    "--color-foreground-lightest"
  );
  const alert = style.getPropertyValue("--color-alert");
  const info = style.getPropertyValue("--color-info");
  const ok = style.getPropertyValue("--color-ok");
  const namedColors = {
    foreground,
    foregroundLightest,
    alert,
    info,
    ok,
  };

  const nodesAndEdges = buildNodesAndEdges(
    props.data,
    props.properties,
    namedColors
  );

  return merge(
    getCommonBaseOptions(),
    getCommonBaseOptionsForGraphType(props.display_type),
    getSeriesForGraphType(
      props.display_type,
      props.data,
      props.properties,
      nodesAndEdges,
      namedColors
    ),
    getOptionOverridesForGraphType(props.display_type, props.properties)
  );
};

const nodeWidth = 70;
const nodeHeight = 70;

const nodeTypes = {
  asset: AssetNode,
};

const edgeTypes = {
  floating: FloatingEdge,
};

interface GraphNode extends Node {}

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
        icon: matchingCategory ? matchingCategory.icon : null,
        label: node.title,
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

  return { nodes, edges };
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

  const nodesAndEdges = buildGraphNodesAndEdges(
    props.data,
    props.properties,
    namedColors
  );
  const [nodes, setNodes, onNodesChange] = useNodesState(nodesAndEdges.nodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(nodesAndEdges.edges);
  // const onConnect = useCallback(
  //   (params) => setEdges((eds) => addEdge(params, eds)),
  //   []
  // );

  return { nodes, edges, onNodesChange, onEdgesChange };
};

const Graph = ({ props, theme, themeWrapperRef }) => {
  const graphOptions = useGraphOptions(props, theme, themeWrapperRef);
  const {} = usePanel();
  return (
    <ReactFlow
      nodes={graphOptions.nodes}
      edges={graphOptions.edges}
      onNodesChange={graphOptions.onNodesChange}
      onEdgesChange={graphOptions.onEdgesChange}
      nodeTypes={nodeTypes}
      // @ts-ignore
      edgeTypes={edgeTypes}
      fitView
      style={{ height: "400px" }}
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
