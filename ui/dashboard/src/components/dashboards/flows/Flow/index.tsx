import ErrorPanel from "../../Error";
import Flows, { FlowProperties, FlowProps, FlowType } from "../index";
import {
  buildNodesAndEdges,
  buildSankeyDataInputs,
  LeafNodeData,
  NodesAndEdges,
  toEChartsType,
} from "../../common";
import { Chart } from "../../charts/Chart";
import { get, merge, set } from "lodash";
import { PanelDefinition, useDashboard } from "../../../../hooks/useDashboard";
import { useEffect, useState } from "react";

const getCommonBaseOptions = () => ({
  animation: false,
  tooltip: {
    trigger: "item",
    triggerOn: "mousemove",
  },
});

const getCommonBaseOptionsForFlowType = (
  type: FlowType = "sankey",
  namedColors
) => {
  switch (type) {
    case "sankey":
      return {};
    default:
      return {};
  }
};

const getSeriesForFlowType = (
  type: FlowType = "sankey",
  data: LeafNodeData | undefined,
  properties: FlowProperties | undefined,
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
      case "sankey": {
        const { data: sankeyData, links } =
          buildSankeyDataInputs(nodesAndEdges);
        series.push({
          type: toEChartsType(type),
          layout: "none",
          draggable: true,
          label: { color: namedColors.foreground, formatter: "{b}" },
          emphasis: {
            focus: "adjacency",
            blurScope: "coordinateSystem",
          },
          lineStyle: {
            color: "source",
            curveness: 0.5,
          },
          data: sankeyData,
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

const getOptionOverridesForFlowType = (
  type: FlowType = "sankey",
  properties: FlowProperties | undefined
) => {
  if (!properties) {
    return {};
  }

  let overrides = {};

  return overrides;
};

const buildFlowOptions = (props: FlowProps, theme, themeWrapperRef) => {
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
    getCommonBaseOptionsForFlowType(props.properties?.type, namedColors),
    getSeriesForFlowType(
      props.properties?.type,
      props.data,
      props.properties,
      nodesAndEdges,
      namedColors
    ),
    getOptionOverridesForFlowType(props.properties?.type, props.properties)
  );
};

const FlowWrapper = (props: FlowProps) => {
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
      options={buildFlowOptions(props, theme, wrapperRef)}
      type={props.properties ? props.properties.type : "sankey"}
    />
  );
};

type FlowDefinition = PanelDefinition & {
  properties: FlowProps;
};

const renderFlow = (definition: FlowDefinition) => {
  // We default to sankey diagram if not specified
  if (!get(definition, "properties.type")) {
    // @ts-ignore
    definition = set(definition, "properties.type", "sankey");
  }
  const {
    properties: { type },
  } = definition;

  const flow = Flows[type];

  if (!flow) {
    return <ErrorPanel error={`Unknown flow type ${type}`} />;
  }

  const Component = flow.component;
  return <Component {...definition} />;
};

const RenderFlow = (props: FlowDefinition) => {
  return renderFlow(props);
};

export default FlowWrapper;

export { RenderFlow };
