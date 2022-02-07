import ErrorPanel from "../../Error";
import Hierarchies, { HierarchyProperties } from "../../hierarchies";
import {
  buildHierarchyDataInputs,
  LeafNodeData,
  toEChartsType,
} from "../../common";
import { Chart } from "../../charts/Chart";
import { HierarchyProps } from "../index";
import { PanelDefinition } from "../../../../hooks/useReport";
import { useEffect, useState } from "react";
import { useTheme } from "../../../../hooks/useTheme";

const getBaseOptions = (type, data, links) => {
  return {
    //tooltip: {
    //    trigger: 'item'
    //},
    series: {
      type: toEChartsType(type),
      layout: "none",
      draggable: true,
      label: { formatter: "{b}" },
      emphasis: {
        focus: "adjacency",
        blurScope: "coordinateSystem",
      },
      lineStyle: {
        color: "source",
        curveness: 0.5,
      },
      data,
      links,
      //data: objectData.map(o => ),
      // categories: Object.entries(categories).map(([category, info]) => ({
      //   name: category,
      //   symbol: "rect",
      //   symbolSize: [160, 40],
      //   itemStyle: { color: info.color },
      // })),
    },
  };
};

const buildHierarchyOptions = (
  type,
  properties: HierarchyProperties,
  data,
  links,
  theme,
  themeWrapperRef
) => {
  return getBaseOptions(type, data, links);
};

const buildHierarchyInputs = (
  rawData: LeafNodeData,
  inputs: HierarchyProperties,
  theme,
  themeWrapperRef
) => {
  if (!rawData) {
    return null;
  }

  const { data, links } = buildHierarchyDataInputs(
    rawData,
    inputs.type,
    inputs
  );
  const options = buildHierarchyOptions(
    inputs.type,
    inputs,
    data,
    links,
    theme,
    themeWrapperRef
  );

  return options;
};

const HierarchyWrapper = ({ data, inputs }) => {
  const [, setRandomVal] = useState(0);
  const { theme, wrapperRef } = useTheme();

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  if (!wrapperRef) {
    return null;
  }

  if (!data) {
    return null;
  }

  return (
    <Chart options={buildHierarchyInputs(data, inputs, theme, wrapperRef)} />
  );
};

type HierarchyDefinition = PanelDefinition & {
  properties: HierarchyProps;
};

const renderChart = (definition: HierarchyDefinition) => {
  // We default to column charts if not specified
  const {
    properties: { type = "sankey" },
  } = definition;
  const hierarchy = Hierarchies[type];

  if (!hierarchy) {
    return <ErrorPanel error={`Unknown chart type ${type}`} />;
  }

  const Component = hierarchy.component;
  return <Component {...definition} />;
};

const RenderHierarchy = (props: HierarchyDefinition) => {
  return renderChart(props);
};

export default HierarchyWrapper;

export { RenderHierarchy };
