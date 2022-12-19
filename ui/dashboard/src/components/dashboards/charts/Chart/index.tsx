import ErrorPanel from "../../Error";
import has from "lodash/has";
import merge from "lodash/merge";
import Placeholder from "../../Placeholder";
import React, { useEffect, useRef, useState } from "react";
import ReactEChartsCore from "echarts-for-react/lib/core";
import set from "lodash/set";
import useChartThemeColors from "../../../../hooks/useChartThemeColors";
import useMediaMode from "../../../../hooks/useMediaMode";
import useTemplateRender from "../../../../hooks/useTemplateRender";
import {
  buildChartDataset,
  getColorOverride,
  LeafNodeData,
  themeColors,
  Width,
} from "../../common";
import { EChartsOption } from "echarts-for-react/src/types";
import {
  ChartProperties,
  ChartProps,
  ChartSeries,
  ChartSeriesOptions,
  ChartTransform,
  ChartType,
} from "../types";
import { FlowType } from "../../flows/types";
import { getChartComponent } from "..";
import { GraphType } from "../../graphs/types";
import { HierarchyType } from "../../hierarchies/types";
import { registerComponent } from "../../index";
import { useDashboard } from "../../../../hooks/useDashboard";
import { useNavigate } from "react-router-dom";

const getThemeColorsWithPointOverrides = (
  type: ChartType = "column",
  series: any[],
  seriesOverrides: ChartSeries | undefined,
  dataset: any[][],
  themeColorValues
) => {
  switch (type) {
    case "donut":
    case "pie": {
      const newThemeColors: string[] = [];
      for (let rowIndex = 1; rowIndex < dataset.length; rowIndex++) {
        if (rowIndex - 1 < themeColors.length) {
          newThemeColors.push(themeColors[rowIndex - 1]);
        } else {
          newThemeColors.push(themeColors[(rowIndex - 1) % themeColors.length]);
        }
      }
      series.forEach((seriesInfo) => {
        const seriesName = seriesInfo.name;
        const overrides = seriesOverrides
          ? seriesOverrides[seriesName] || {}
          : ({} as ChartSeriesOptions);
        const pointOverrides = overrides.points || {};
        dataset.slice(1).forEach((dataRow, dataRowIndex) => {
          const pointOverride = pointOverrides[dataRow[0]];
          if (pointOverride && pointOverride.color) {
            newThemeColors[dataRowIndex] = getColorOverride(
              pointOverride.color,
              themeColorValues
            );
          }
        });
      });
      return newThemeColors;
    }
    default:
      const newThemeColors: string[] = [];
      for (let seriesIndex = 0; seriesIndex < series.length; seriesIndex++) {
        if (seriesIndex < themeColors.length - 1) {
          newThemeColors.push(themeColors[seriesIndex]);
        } else {
          newThemeColors.push(themeColors[seriesIndex % themeColors.length]);
        }
      }
      return newThemeColors;
  }
};

const getCommonBaseOptions = () => ({
  animation: false,
  grid: {
    bottom: 40,
    containLabel: true,
  },
  legend: {
    orient: "horizontal",
    left: "center",
    top: "10",
    textStyle: {
      fontSize: 11,
    },
  },
  textStyle: {
    fontFamily:
      'ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"',
  },
  tooltip: {
    appendToBody: true,
    textStyle: {
      fontSize: 11,
    },
    trigger: "item",
  },
});

const getXAxisLabelRotation = (number_of_rows: number) => {
  if (number_of_rows < 5) {
    return 0;
  }
  if (number_of_rows < 10) {
    return 30;
  }
  if (number_of_rows < 15) {
    return 45;
  }
  if (number_of_rows < 20) {
    return 60;
  }
  return 90;
};

const getXAxisLabelWidth = (number_of_rows: number) => {
  if (number_of_rows < 5) {
    return null;
  }
  if (number_of_rows < 10) {
    return 85;
  }
  if (number_of_rows < 15) {
    return 75;
  }
  if (number_of_rows < 20) {
    return 60;
  }
  return 50;
};

const getCommonBaseOptionsForChartType = (
  type: ChartType | undefined,
  width: Width | undefined,
  dataset: any[][],
  series: any[],
  seriesOverrides: ChartSeries | undefined,
  themeColors
) => {
  switch (type) {
    case "bar":
      return {
        color: getThemeColorsWithPointOverrides(
          type,
          series,
          seriesOverrides,
          dataset,
          themeColors
        ),
        legend: {
          show: series ? series.length > 1 : false,
          textStyle: {
            color: themeColors.foreground,
          },
        },
        // Declare an x-axis (category axis).
        // The category map the first row in the dataset by default.
        xAxis: {
          axisLabel: { color: themeColors.foreground },
          axisLine: {
            show: true,
            lineStyle: { color: themeColors.foregroundLightest },
          },
          axisTick: { show: true },
          nameGap: 30,
          nameLocation: "center",
          nameTextStyle: { color: themeColors.foreground },
          splitLine: { show: false },
        },
        // Declare a y-axis (value axis).
        yAxis: {
          type: "category",
          axisLabel: {
            color: themeColors.foreground,
            width: 50,
            overflow: "truncate",
          },
          axisLine: { lineStyle: { color: themeColors.foregroundLightest } },
          axisTick: { show: false },
          nameGap: width ? width + 42 : 50,
          nameLocation: "center",
          nameTextStyle: { color: themeColors.foreground },
        },
      };
    case "area":
    case "line":
      return {
        color: getThemeColorsWithPointOverrides(
          type,
          series,
          seriesOverrides,
          dataset,
          themeColors
        ),
        legend: {
          show: series ? series.length > 1 : false,
          textStyle: {
            color: themeColors.foreground,
          },
        },
        // Declare an x-axis (category axis).
        // The category map the first row in the dataset by default.
        xAxis: {
          type: "category",
          boundaryGap: type !== "area",
          axisLabel: {
            color: themeColors.foreground,
            rotate: getXAxisLabelRotation(dataset.length - 1),
            width: getXAxisLabelWidth(dataset.length),
            overflow: "truncate",
          },
          axisLine: { lineStyle: { color: themeColors.foregroundLightest } },
          axisTick: { show: false },
          nameGap: 30,
          nameLocation: "center",
          nameTextStyle: { color: themeColors.foreground },
        },
        // Declare a y-axis (value axis).
        yAxis: {
          axisLabel: { color: themeColors.foreground },
          axisLine: {
            show: true,
            lineStyle: { color: themeColors.foregroundLightest },
          },
          axisTick: { show: true },
          splitLine: { show: false },
          nameGap: width ? width + 42 : 50,
          nameLocation: "center",
          nameTextStyle: { color: themeColors.foreground },
        },
        tooltip: {
          trigger: "axis",
        },
      };
    case "column":
      return {
        color: getThemeColorsWithPointOverrides(
          type,
          series,
          seriesOverrides,
          dataset,
          themeColors
        ),
        legend: {
          show: series ? series.length > 1 : false,
          textStyle: {
            color: themeColors.foreground,
          },
        },
        // Declare an x-axis (category axis).
        // The category map the first row in the dataset by default.
        xAxis: {
          type: "category",
          axisLabel: {
            color: themeColors.foreground,
            rotate: getXAxisLabelRotation(dataset.length - 1),
            width: getXAxisLabelWidth(dataset.length),
            overflow: "truncate",
          },
          axisLine: { lineStyle: { color: themeColors.foregroundLightest } },
          axisTick: { show: false },
          nameGap: 30,
          nameLocation: "center",
          nameTextStyle: { color: themeColors.foreground },
        },
        // Declare a y-axis (value axis).
        yAxis: {
          axisLabel: { color: themeColors.foreground },
          axisLine: {
            show: true,
            lineStyle: { color: themeColors.foregroundLightest },
          },
          axisTick: { show: true },
          splitLine: { show: false },
          nameGap: width ? width + 42 : 50,
          nameLocation: "center",
          nameTextStyle: { color: themeColors.foreground },
        },
      };
    case "pie":
      return {
        color: getThemeColorsWithPointOverrides(
          type,
          series,
          seriesOverrides,
          dataset,
          themeColors
        ),
        legend: {
          show: false,
          textStyle: {
            color: themeColors.foreground,
          },
        },
      };
    case "donut":
      return {
        color: getThemeColorsWithPointOverrides(
          type,
          series,
          seriesOverrides,
          dataset,
          themeColors
        ),
        legend: {
          show: false,
          textStyle: {
            color: themeColors.foreground,
          },
        },
      };
    default:
      return {};
  }
};

const getOptionOverridesForChartType = (
  type: ChartType = "column",
  properties: ChartProperties | undefined
) => {
  if (!properties) {
    return {};
  }

  let overrides = {};

  // orient: "horizontal",
  //     left: "center",
  //     top: "top",

  if (properties.legend) {
    // Legend display
    const legendDisplay = properties.legend.display;
    if (legendDisplay === "all") {
      overrides = set(overrides, "legend.show", true);
    } else if (legendDisplay === "none") {
      overrides = set(overrides, "legend.show", false);
    }

    // Legend display position
    const legendPosition = properties.legend.position;
    if (legendPosition === "top") {
      overrides = set(overrides, "legend.orient", "horizontal");
      overrides = set(overrides, "legend.left", "center");
      overrides = set(overrides, "legend.top", 10);
      overrides = set(overrides, "legend.bottom", "auto");
    } else if (legendPosition === "right") {
      overrides = set(overrides, "legend.orient", "vertical");
      overrides = set(overrides, "legend.left", "right");
      overrides = set(overrides, "legend.top", "middle");
      overrides = set(overrides, "legend.bottom", "auto");
      overrides = set(overrides, "grid.right", "20%");
    } else if (legendPosition === "bottom") {
      overrides = set(overrides, "legend.orient", "horizontal");
      overrides = set(overrides, "legend.left", "center");
      overrides = set(overrides, "legend.top", "auto");
      overrides = set(overrides, "legend.bottom", 10);
      overrides = set(overrides, "grid.top", 30);
    } else if (legendPosition === "left") {
      overrides = set(overrides, "legend.orient", "vertical");
      overrides = set(overrides, "legend.left", "left");
      overrides = set(overrides, "legend.top", "middle");
      overrides = set(overrides, "legend.bottom", "auto");
      overrides = set(overrides, "grid.left", "20%");
    }
  }

  // Axes settings
  if (properties.axes) {
    // X axis settings
    if (properties.axes.x) {
      // X axis display setting
      const xAxisDisplay = properties.axes.x.display;
      if (xAxisDisplay === "all") {
        overrides = set(overrides, "xAxis.show", true);
      } else if (xAxisDisplay === "none") {
        overrides = set(overrides, "xAxis.show", false);
      }

      // X axis min setting
      if (type === "bar" && has(properties, "axes.x.min")) {
        overrides = set(overrides, "xAxis.min", properties.axes.x.min);
      }
      // Y axis max setting
      if (type === "bar" && has(properties, "axes.x.max")) {
        overrides = set(overrides, "xAxis.max", properties.axes.x.max);
      }

      // X axis labels settings
      if (properties.axes.x.labels) {
        // X axis labels display setting
        const xAxisTicksDisplay = properties.axes.x.labels.display;
        if (xAxisTicksDisplay === "all") {
          overrides = set(overrides, "xAxis.axisLabel.show", true);
        } else if (xAxisTicksDisplay === "none") {
          overrides = set(overrides, "xAxis.axisLabel.show", false);
        }
      }

      // X axis title settings
      if (properties.axes.x.title) {
        // X axis title display setting
        const xAxisTitleDisplay = properties.axes.x.title.display;
        if (xAxisTitleDisplay === "none") {
          overrides = set(overrides, "xAxis.name", null);
        }

        // X Axis title align setting
        const xAxisTitleAlign = properties.axes.x.title.align;
        if (xAxisTitleAlign === "start") {
          overrides = set(overrides, "xAxis.nameLocation", "start");
        } else if (xAxisTitleAlign === "center") {
          overrides = set(overrides, "xAxis.nameLocation", "center");
        } else if (xAxisTitleAlign === "end") {
          overrides = set(overrides, "xAxis.nameLocation", "end");
        }

        // X Axis title value setting
        const xAxisTitleValue = properties.axes.x.title.value;
        if (xAxisTitleValue) {
          overrides = set(overrides, "xAxis.name", xAxisTitleValue);
        }
      }
    }

    // Y axis settings
    if (properties.axes.y) {
      // Y axis display setting
      const yAxisDisplay = properties.axes.y.display;
      if (yAxisDisplay === "all") {
        overrides = set(overrides, "yAxis.show", true);
      } else if (yAxisDisplay === "none") {
        overrides = set(overrides, "yAxis.show", false);
      }

      // Y axis min setting
      if (type !== "bar" && has(properties, "axes.y.min")) {
        overrides = set(overrides, "yAxis.min", properties.axes.y.min);
      }
      // Y axis max setting
      if (type !== "bar" && has(properties, "axes.y.max")) {
        overrides = set(overrides, "yAxis.max", properties.axes.y.max);
      }

      // Y axis labels settings
      if (properties.axes.y.labels) {
        // Y axis labels display setting
        const yAxisTicksDisplay = properties.axes.y.labels.display;
        if (yAxisTicksDisplay === "all") {
          overrides = set(overrides, "yAxis.axisLabel.show", true);
        } else if (yAxisTicksDisplay === "none") {
          overrides = set(overrides, "yAxis.axisLabel.show", false);
        }
      }

      // Y axis title settings
      if (properties.axes.y.title) {
        // Y axis title display setting
        const yAxisTitleDisplay = properties.axes.y.title.display;
        if (yAxisTitleDisplay === "none") {
          overrides = set(overrides, "yAxis.name", null);
        }

        // Y Axis title align setting
        const yAxisTitleAlign = properties.axes.y.title.align;
        if (yAxisTitleAlign === "start") {
          overrides = set(overrides, "yAxis.nameLocation", "start");
        } else if (yAxisTitleAlign === "center") {
          overrides = set(overrides, "yAxis.nameLocation", "center");
        } else if (yAxisTitleAlign === "end") {
          overrides = set(overrides, "yAxis.nameLocation", "end");
        }

        // Y Axis title value setting
        const yAxisTitleValue = properties.axes.y.title.value;
        if (yAxisTitleValue) {
          overrides = set(overrides, "yAxis.name", yAxisTitleValue);
        }
      }
    }
  }

  return overrides;
};

const getSeriesForChartType = (
  type: ChartType = "column",
  data: LeafNodeData | undefined,
  properties: ChartProperties | undefined,
  rowSeriesLabels: string[],
  transform: ChartTransform,
  themeColors
) => {
  if (!data) {
    return [];
  }
  const series: any[] = [];
  const seriesNames =
    transform === "crosstab"
      ? rowSeriesLabels
      : data.columns.slice(1).map((col) => col.name);
  const seriesLength = seriesNames.length;
  for (let seriesIndex = 0; seriesIndex < seriesLength; seriesIndex++) {
    let seriesName = seriesNames[seriesIndex];
    let seriesColor = "auto";
    let seriesOverrides;
    if (properties) {
      if (properties.series && properties.series[seriesName]) {
        seriesOverrides = properties.series[seriesName];
      }
      if (seriesOverrides && seriesOverrides.title) {
        seriesName = seriesOverrides.title;
      }
      if (seriesOverrides && seriesOverrides.color) {
        seriesColor = getColorOverride(seriesOverrides.color, themeColors);
      }
    }

    switch (type) {
      case "bar":
      case "column":
        series.push({
          name: seriesName,
          type: "bar",
          ...(properties && properties.grouping === "compare"
            ? {}
            : { stack: "total" }),
          itemStyle: { color: seriesColor },
          // label: {
          //   show: true,
          //   position: 'outside'
          // },
        });
        break;
      case "donut":
        series.push({
          name: seriesName,
          type: "pie",
          center: ["50%", "45%"],
          radius: ["30%", "50%"],
          label: { color: themeColors.foreground },
        });
        break;
      case "area":
        series.push({
          name: seriesName,
          type: "line",
          ...(properties && properties.grouping === "compare"
            ? {}
            : { stack: "total" }),
          areaStyle: {},
          emphasis: {
            focus: "series",
          },
          itemStyle: { color: seriesColor },
        });
        break;
      case "line":
        series.push({
          name: seriesName,
          type: "line",
          itemStyle: { color: seriesColor },
        });
        break;
      case "pie":
        series.push({
          name: seriesName,
          type: "pie",
          center: ["50%", "40%"],
          radius: "50%",
          label: { color: themeColors.foreground },
          emphasis: {
            itemStyle: {
              shadowBlur: 5,
              shadowOffsetX: 0,
              shadowColor: "rgba(0, 0, 0, 0.5)",
            },
          },
        });
    }
  }
  return series;
};

const buildChartOptions = (props: ChartProps, themeColors: any) => {
  const { dataset, rowSeriesLabels, transform } = buildChartDataset(
    props.data,
    props.properties
  );
  const series = getSeriesForChartType(
    props.display_type || "column",
    props.data,
    props.properties,
    rowSeriesLabels,
    transform,
    themeColors
  );
  return merge(
    getCommonBaseOptions(),
    getCommonBaseOptionsForChartType(
      props.display_type || "column",
      props.width,
      dataset,
      series,
      props.properties?.series,
      themeColors
    ),
    getOptionOverridesForChartType(
      props.display_type || "column",
      props.properties
    ),
    { series },
    {
      dataset: {
        source: dataset,
      },
    }
  );
};

type ChartComponentProps = {
  options: EChartsOption;
  type: ChartType | FlowType | GraphType | HierarchyType;
};

const handleClick = async (params: any, navigate, renderTemplates) => {
  const componentType = params.componentType;
  if (componentType !== "series") {
    return;
  }
  const dataType = params.dataType;

  switch (dataType) {
    case "node":
      if (!params.data.href) {
        return;
      }
      const renderedResults = await renderTemplates(
        { graph_node: params.data.href as string },
        [params.data]
      );
      let rowRenderResult = renderedResults[0];
      navigate(rowRenderResult.graph_node.result);
  }
};

const Chart = ({ options, type }: ChartComponentProps) => {
  const [echarts, setEcharts] = useState<any | null>(null);
  const navigate = useNavigate();
  const chartRef = useRef<ReactEChartsCore>(null);
  const [imageUrl, setImageUrl] = useState<string | null>(null);
  const mediaMode = useMediaMode();
  const { ready: templateRenderReady, renderTemplates } = useTemplateRender();

  // Dynamically import echarts from its own bundle
  useEffect(() => {
    import("./echarts").then((m) => setEcharts(m.echarts));
  }, []);

  useEffect(() => {
    if (!chartRef.current || !options) {
      return;
    }

    const echartInstance = chartRef.current.getEchartsInstance();
    const dataURL = echartInstance.getDataURL({});
    if (dataURL === imageUrl) {
      return;
    }
    setImageUrl(dataURL);
  }, [chartRef, imageUrl, options]);

  if (!options) {
    return null;
  }

  const eventsDict = {
    click: (params) => handleClick(params, navigate, renderTemplates),
  };

  const PlaceholderComponent = Placeholder.component;

  return (
    <PlaceholderComponent ready={!!echarts && templateRenderReady}>
      <>
        {mediaMode !== "print" && (
          <div className="relative">
            <ReactEChartsCore
              ref={chartRef}
              echarts={echarts}
              className="chart-canvas"
              onEvents={eventsDict}
              option={options}
              notMerge={true}
              lazyUpdate={true}
              style={
                type === "pie" || type === "donut" ? { height: "250px" } : {}
              }
            />
          </div>
        )}
        {mediaMode === "print" && imageUrl && (
          <div>
            <img alt="Chart" className="max-w-full max-h-full" src={imageUrl} />
          </div>
        )}
      </>
    </PlaceholderComponent>
  );
};

const ChartWrapper = (props: ChartProps) => {
  const {
    themeContext: { wrapperRef },
  } = useDashboard();
  const themeColors = useChartThemeColors();

  if (!wrapperRef) {
    return null;
  }

  if (!props.data) {
    return null;
  }

  return (
    <Chart
      options={buildChartOptions(props, themeColors)}
      type={props.display_type || "column"}
    />
  );
};

const renderChart = (definition: ChartProps) => {
  // We default to column charts if not specified
  const { display_type = "column" } = definition;

  const chart = getChartComponent(display_type);

  if (!chart) {
    return <ErrorPanel error={`Unknown chart type ${display_type}`} />;
  }

  const Component = chart.component;
  return <Component {...definition} />;
};

const RenderChart = (props: ChartProps) => {
  return renderChart(props);
};

registerComponent("chart", RenderChart);

export default ChartWrapper;

export { Chart };
