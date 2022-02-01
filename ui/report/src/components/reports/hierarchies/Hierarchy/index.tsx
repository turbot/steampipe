import ErrorPanel from "../../Error";
import Hierarchies, { HierarchyProperties } from "../../hierarchies";
import ReactEChartsCore from "echarts-for-react/lib/core";
import * as echarts from "echarts/core";
import { buildHierarchyDataInputs, LeafNodeData } from "../../common";
import { CanvasRenderer } from "echarts/renderers";
import {
  DatasetComponent,
  GridComponent,
  TitleComponent,
  TooltipComponent,
} from "echarts/components";
import { EChartsType, HierarchyProps, HierarchyType } from "../index";
import { PanelDefinition } from "../../../../hooks/useReport";
import { SankeyChart } from "echarts/charts";
import { useTheme } from "../../../../hooks/useTheme";

echarts.use([
  CanvasRenderer,
  DatasetComponent,
  GridComponent,
  SankeyChart,
  TitleComponent,
  TooltipComponent,
]);

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

const toEChartsType = (type: HierarchyType): EChartsType => {
  return type as EChartsType;
};

const Hierarchy = ({ data, inputs, theme, themeWrapperRef }) => {
  const options = buildHierarchyInputs(data, inputs, theme, themeWrapperRef);
  // console.log(built);
  // const options = {
  //   series: {
  //     type: "sankey",
  //     layout: "none",
  //     emphasis: {
  //       focus: "adjacency",
  //     },
  //     data: [
  //       {
  //         name: "a",
  //       },
  //       {
  //         name: "b",
  //       },
  //       {
  //         name: "a1",
  //       },
  //       {
  //         name: "a2",
  //       },
  //       {
  //         name: "b1",
  //       },
  //       {
  //         name: "c",
  //       },
  //     ],
  //     links: [
  //       {
  //         source: "a",
  //         target: "a1",
  //         value: 5,
  //       },
  //       {
  //         source: "a",
  //         target: "a2",
  //         value: 3,
  //       },
  //       {
  //         source: "b",
  //         target: "b1",
  //         value: 8,
  //       },
  //       {
  //         source: "a",
  //         target: "b1",
  //         value: 3,
  //       },
  //       {
  //         source: "b1",
  //         target: "a1",
  //         value: 1,
  //       },
  //       {
  //         source: "b1",
  //         target: "c",
  //         value: 2,
  //       },
  //     ],
  //   },
  // };
  return (
    <ReactEChartsCore
      echarts={echarts}
      option={options}
      notMerge={true}
      lazyUpdate={true}
      theme={"theme_name"}
      // onChartReady={this.onChartReadyCallback}
      // onEvents={EventsDict}
      // opts={{}}
      showLoading={false}
    />
  );
};

const HierarchyWrapper = ({ data, inputs }) => {
  const { theme, wrapperRef } = useTheme();

  if (!wrapperRef) {
    return null;
  }

  if (!data) {
    return null;
  }

  return (
    <Hierarchy
      data={data}
      inputs={inputs}
      theme={theme}
      themeWrapperRef={wrapperRef}
    />
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
