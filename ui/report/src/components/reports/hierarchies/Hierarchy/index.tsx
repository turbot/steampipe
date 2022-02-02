import ErrorPanel from "../../Error";
import Hierarchies, { HierarchyProperties } from "../../hierarchies";
import ReactEChartsCore from "echarts-for-react/lib/core";
import useMediaMode from "../../../../hooks/useMediaMode";
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
import { PanelDefinition, useReport } from "../../../../hooks/useReport";
import { SankeyChart } from "echarts/charts";
import { usePanel } from "../../../../hooks/usePanel";
import { useEffect, useRef, useState } from "react";
import { useTheme } from "../../../../hooks/useTheme";
import { ZoomIcon } from "../../../../constants/icons";

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
  const [_, setRandomVal] = useState(0);
  const chartRef = useRef<ReactEChartsCore>(null);
  const [imageUrl, setImageUrl] = useState<string | null>(null);
  const [showZoom, setShowZoom] = useState(false);
  const { definition: panelDefinition, showExpand } = usePanel();
  const { dispatch } = useReport();
  const mediaMode = useMediaMode();

  const options = buildHierarchyInputs(data, inputs, theme, themeWrapperRef);

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  useEffect(() => {
    if (!chartRef.current || !options) {
      return;
    }

    const echartInstance = chartRef.current.getEchartsInstance();
    setImageUrl(echartInstance.getDataURL());
  }, [chartRef, inputs, options]);

  return (
    <>
      {mediaMode !== "print" && (
        <div
          className="relative"
          onMouseEnter={() => {
            if (!showExpand) {
              return;
            }
            setShowZoom(true);
          }}
          onMouseLeave={() => {
            if (!showExpand) {
              return;
            }
            setShowZoom(false);
          }}
        >
          {showZoom && (
            <div
              className="absolute right-0 top-0 cursor-pointer z-50"
              onClick={() =>
                dispatch({ type: "select_panel", panel: panelDefinition })
              }
            >
              <ZoomIcon className="h-5 w-5 text-black-scale-4" />
            </div>
          )}
          <ReactEChartsCore
            ref={chartRef}
            echarts={echarts}
            option={options}
            notMerge={true}
            lazyUpdate={true}
          />
        </div>
      )}
      {mediaMode === "print" && imageUrl && (
        <div>
          <img className="max-w-full max-h-full" src={imageUrl} />
        </div>
      )}
    </>
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
