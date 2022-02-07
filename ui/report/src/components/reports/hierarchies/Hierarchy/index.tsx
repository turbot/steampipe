import ErrorPanel from "../../Error";
import Hierarchies, {
  HierarchyProperties,
  HierarchyType,
} from "../../hierarchies";
import {
  buildSankeyDataInputs,
  buildTreeDataInputs,
  LeafNodeData,
  toEChartsType,
} from "../../common";
import { Chart } from "../../charts/Chart";
import { HierarchyProps } from "../index";
import { merge } from "lodash";
import { PanelDefinition } from "../../../../hooks/useReport";
import { useEffect, useState } from "react";
import { useTheme } from "../../../../hooks/useTheme";

// const getBaseOptions = (type, data, links) => {
//   return {
//     //tooltip: {
//     //    trigger: 'item'
//     //},
//     series: {
//       type: toEChartsType(type),
//       layout: "none",
//       draggable: true,
//       label: { formatter: "{b}" },
//       emphasis: {
//         focus: "adjacency",
//         blurScope: "coordinateSystem",
//       },
//       lineStyle: {
//         color: "source",
//         curveness: 0.5,
//       },
//       data,
//       links,
//       //data: objectData.map(o => ),
//       // categories: Object.entries(categories).map(([category, info]) => ({
//       //   name: category,
//       //   symbol: "rect",
//       //   symbolSize: [160, 40],
//       //   itemStyle: { color: info.color },
//       // })),
//     },
//   };
// };

const getCommonBaseOptions = () => ({
  tooltip: {
    trigger: "item",
    triggerOn: "mousemove",
  },
});

const getCommonBaseOptionsForHierarchyType = (
  type: HierarchyType = "sankey",
  themeColors
) => {
  switch (type) {
    case "sankey":
      return {};
    default:
      return {};
  }
};

const getSeriesForHierarchyType = (
  type: HierarchyType = "sankey",
  data: LeafNodeData | undefined,
  properties: HierarchyProperties | undefined,
  themeColors
) => {
  if (!data) {
    return {};
  }
  const series: any[] = [];
  const seriesLength = 1;
  for (let seriesIndex = 0; seriesIndex < seriesLength; seriesIndex++) {
    switch (type) {
      case "sankey": {
        const { data: sankeyData, links } = buildSankeyDataInputs(
          data,
          properties
        );
        series.push({
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
          data: sankeyData,
          links,
        });
        break;
      }
      case "tree": {
        const { data: treeData } = buildTreeDataInputs(data, properties);
        series.push({
          type: "tree",
          data: treeData,
          top: "1%",
          left: "7%",
          bottom: "1%",
          right: "20%",
          symbolSize: 7,
          label: {
            position: "left",
            verticalAlign: "middle",
            align: "right",
            fontSize: 9,
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
  type: HierarchyType = "sankey",
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
  const themeColors = {
    foreground,
    foregroundLightest,
  };
  return merge(
    getCommonBaseOptions(),
    getCommonBaseOptionsForHierarchyType(props.properties?.type, themeColors),
    getSeriesForHierarchyType(
      props.properties?.type,
      props.data,
      props.properties,
      themeColors
    ),
    getOptionOverridesForHierarchyType(props.properties?.type, props.properties)
  );
};

const HierarchyWrapper = (props: HierarchyProps) => {
  const [, setRandomVal] = useState(0);
  const { theme, wrapperRef } = useTheme();

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  if (!wrapperRef) {
    return null;
  }

  if (!props.data) {
    return null;
  }

  return <Chart options={buildHierarchyOptions(props, theme, wrapperRef)} />;
};

type HierarchyDefinition = PanelDefinition & {
  properties: HierarchyProps;
};

const renderHierarchy = (definition: HierarchyDefinition) => {
  // We default to sankey diagram if not specified
  const {
    properties: { type = "sankey" },
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
