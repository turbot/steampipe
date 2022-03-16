import ErrorPanel from "../../Error";
import Hierarchies, {
  HierarchyProperties,
  HierarchyType,
} from "../../hierarchies";
import {
  buildNodesAndEdges,
  buildTreeDataInputs,
  LeafNodeData,
  NodesAndEdges,
} from "../../common";
import { Chart } from "../../charts/Chart";
import { get, merge, set } from "lodash";
import { HierarchyProps } from "../index";
import { PanelDefinition, useDashboard } from "../../../../hooks/useDashboard";
import { useEffect, useState } from "react";

const getCommonBaseOptions = () => ({
  animation: false,
  tooltip: {
    trigger: "item",
    triggerOn: "mousemove",
  },
});

const getCommonBaseOptionsForHierarchyType = (
  type: HierarchyType = "tree",
  namedColors
) => {
  switch (type) {
    default:
      return {};
  }
};

const getSeriesForHierarchyType = (
  type: HierarchyType = "tree",
  data: LeafNodeData | undefined,
  properties: HierarchyProperties | undefined,
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
      case "tree": {
        const { data: treeData } = buildTreeDataInputs(nodesAndEdges);
        series.push({
          type: "tree",
          data: treeData,
          top: "1%",
          left: "7%",
          bottom: "1%",
          right: "20%",
          symbolSize: 7,
          label: {
            color: namedColors.foreground,
            position: "left",
            verticalAlign: "middle",
            align: "right",
          },
          leaves: {
            label: {
              position: "right",
              verticalAlign: "middle",
              align: "left",
            },
          },
          emphasis: {
            focus: "descendant",
          },
          expandAndCollapse: false,
          animationDuration: 550,
          animationDurationUpdate: 750,
        });
      }
    }
  }

  return { series };
};

const getOptionOverridesForHierarchyType = (
  type: HierarchyType = "tree",
  properties: HierarchyProperties | undefined
) => {
  if (!properties) {
    return {};
  }

  let overrides = {};

  return overrides;
};

const buildHierarchyOptions = (
  props: HierarchyProps,
  theme,
  themeWrapperRef
) => {
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
    getCommonBaseOptionsForHierarchyType(props.properties?.type, namedColors),
    getSeriesForHierarchyType(
      props.properties?.type,
      props.data,
      props.properties,
      nodesAndEdges,
      namedColors
    ),
    getOptionOverridesForHierarchyType(props.properties?.type, props.properties)
  );
};

const HierarchyWrapper = (props: HierarchyProps) => {
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
      options={buildHierarchyOptions(props, theme, wrapperRef)}
      type={props.properties ? props.properties.type : "tree"}
    />
  );
};

type HierarchyDefinition = PanelDefinition & {
  properties: HierarchyProps;
};

const renderHierarchy = (definition: HierarchyDefinition) => {
  // We default to tree diagram if not specified
  if (!get(definition, "properties.type")) {
    // @ts-ignore
    definition = set(definition, "properties.type", "tree");
  }
  const {
    properties: { type },
  } = definition;

  const hierarchy = Hierarchies[type];

  if (!hierarchy) {
    return <ErrorPanel error={`Unknown hierarchy type ${type}`} />;
  }

  const Component = hierarchy.component;
  return <Component {...definition} />;
};

const RenderHierarchy = (props: HierarchyDefinition) => {
  return renderHierarchy(props);
};

export default HierarchyWrapper;

export { RenderHierarchy };
