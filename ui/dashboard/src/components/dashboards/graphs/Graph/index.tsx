import ErrorPanel from "../../Error";
import merge from "lodash/merge";
import {
  buildGraphDataInputs,
  buildNodesAndEdges,
  LeafNodeData,
  NodesAndEdges,
  toEChartsType,
} from "../../common";
import { Chart } from "../../charts/Chart";
import { GraphProperties, GraphProps, GraphType } from "../types";
import { getGraphComponent } from "..";
import { registerComponent } from "../../index";
import { useDashboard } from "../../../../hooks/useDashboard";
import { useEffect, useState } from "react";

const getCommonBaseOptions = () => ({
  animation: false,
  tooltip: {
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
            formatter: "{b}",
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

const GraphWrapper = (props: GraphProps) => {
  const [, setRandomVal] = useState(0);
  const {
    themeContext: { theme, wrapperRef },
  } = useDashboard();

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  if (!wrapperRef) {
    return null;
  }

  if (!props.data) {
    return null;
  }

  return (
    <Chart
      options={buildGraphOptions(props, theme, wrapperRef)}
      type={props.display_type || "graph"}
    />
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
