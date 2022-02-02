import Charts, { ChartProperties, ChartProps, ChartType } from "../index";
import ErrorPanel from "../../Error";
import ReactEChartsCore from "echarts-for-react/lib/core";
import useMediaMode from "../../../../hooks/useMediaMode";
import { BarChart, LineChart, PieChart, SankeyChart } from "echarts/charts";
import { buildChartDataset, LeafNodeData, themeColors } from "../../common";
import { CanvasRenderer } from "echarts/renderers";
import {
  DatasetComponent,
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent,
} from "echarts/components";
import { EChartsOption } from "echarts-for-react/src/types";
import { LabelLayout } from "echarts/features";
import { merge } from "lodash";
import { PanelDefinition, useReport } from "../../../../hooks/useReport";
import { Theme, useTheme } from "../../../../hooks/useTheme";
import { useEffect, useRef, useState } from "react";
import { usePanel } from "../../../../hooks/usePanel";
import { ZoomIcon } from "../../../../constants/icons";
import * as echarts from "echarts/core";

echarts.use([
  BarChart,
  LabelLayout,
  LegendComponent,
  LineChart,
  PieChart,
  CanvasRenderer,
  DatasetComponent,
  GridComponent,
  SankeyChart,
  TitleComponent,
  TooltipComponent,
]);

const getCommonBaseOptions = () => ({
  animation: false,
  color: themeColors,
  legend: {
    orient: "horizontal",
    textStyle: {
      fontSize: 11,
    },
  },
  tooltip: {
    trigger: "item",
  },
});

const getCommonBaseOptionsForChartType = (
  type: ChartType | undefined,
  series: any[] | undefined,
  themeColors
) => {
  switch (type) {
    case "bar":
      return {
        legend: {
          show: series ? series.length > 1 : false,
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
          splitLine: { show: false },
        },
        // Declare a y-axis (value axis).
        yAxis: {
          type: "category",
          axisLabel: { color: themeColors.foreground },
          axisLine: { lineStyle: { color: themeColors.foregroundLightest } },
          axisTick: { show: false },
        },
      };
    case "column":
    case "line":
      return {
        legend: {
          show: series ? series.length > 1 : false,
        },
        // Declare an x-axis (category axis).
        // The category map the first row in the dataset by default.
        xAxis: {
          type: "category",
          axisLabel: { color: themeColors.foreground },
          axisLine: { lineStyle: { color: themeColors.foregroundLightest } },
          axisTick: { show: false },
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
        },
      };
    case "pie":
      return {
        legend: {
          show: true,
        },
        series: [
          {
            type: "pie",
            radius: "50%",
            emphasis: {
              itemStyle: {
                shadowBlur: 5,
                shadowOffsetX: 0,
                shadowColor: "rgba(0, 0, 0, 0.5)",
              },
            },
          },
        ],
      };
    case "donut":
      return {
        legend: {
          show: true,
        },
      };
    default:
      return {};
  }
};

const getOptionOverridesForChartType = (
  properties: ChartProperties | undefined
) => {
  if (!properties) {
    return {};
  }
};

const getSeriesForChartType = (
  type: ChartType = "column",
  data: LeafNodeData | undefined,
  properties: ChartProperties | undefined
) => {
  if (!data) {
    return {};
  }
  const series: any[] = [];
  const seriesLength = data.columns.length - 1;
  for (let seriesIndex = 1; seriesIndex <= seriesLength; seriesIndex++) {
    let seriesName = data.columns[seriesIndex].name;
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
        seriesColor = seriesOverrides.color;
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
        });
        break;
      case "donut":
        series.push({ name: seriesName, type: "pie", radius: ["40%", "70%"] });
        break;
      case "line":
        series.push({
          name: seriesName,
          type: "line",
          itemStyle: { color: seriesColor },
        });
        break;
    }
  }
  return { series };
};

const buildChartOptions = (
  props: ChartProps,
  theme: Theme,
  themeWrapperRef: ((instance: null) => void) | React.RefObject<null>
) => {
  // We need to get the theme CSS variable values - these are accessible on the theme root element and below in the tree
  // @ts-ignore
  const style = window.getComputedStyle(themeWrapperRef);
  const foreground = style.getPropertyValue("--color-foreground");
  const foregroundLightest = style.getPropertyValue(
    "--color-foreground-lightest"
  );
  const seriesData = getSeriesForChartType(
    props.properties?.type,
    props.data,
    props.properties
  );
  return merge(
    getCommonBaseOptions(),
    getCommonBaseOptionsForChartType(
      props.properties?.type,
      seriesData.series,
      {
        foreground,
        foregroundLightest,
      }
    ),
    getOptionOverridesForChartType(props.properties),
    seriesData,
    {
      dataset: {
        source: buildChartDataset(props.data),
      },
    }
  );
};

const getBaseOptions = (type, data, min, max, theme, themeWrapperRef) => {
  // We need to get the theme CSS variable values - these are accessible on the theme root element and below in the tree
  // @ts-ignore
  const style = window.getComputedStyle(themeWrapperRef);
  const foreground = style.getPropertyValue("--color-foreground");
  const foregroundLightest = style.getPropertyValue(
    "--color-foreground-lightest"
  );

  return {
    tooltip: {
      trigger: "item",
    },
    legend: {
      show: type === "donut" || type === "pie" || data.series.length > 1,
      // position: type === "donut" || type === "pie" ? "right" : "top",
      orient: "vertical",
      left: "left",
    },
  };

  return {
    responsive: true,
    //maintainAspectRatio: false,
    layout: {
      padding: 5,
    },
    animation: {
      duration: 0,
    },
    datasets: {
      line: {
        borderColor: (context) => context.dataset.backgroundColor,
        borderWidth: 1,
      },
    },
    // Bar charts should be horizontal
    indexAxis: type === "bar" ? "y" : "x",
    scales:
      type !== "donut" && type !== "pie"
        ? {
            x: {
              ticks: {
                color: foreground,
              },
              grid: {
                borderColor: foregroundLightest,
                display: false,
                color: foregroundLightest,
              },
              stacked: true,
            },

            y: {
              suggestedMin: min,
              suggestedMax: max,
              ticks: {
                color: foreground,
              },
              grid: {
                borderColor: foregroundLightest,
                display: false,
                color: foregroundLightest,
              },
              stacked: true,
            },
          }
        : {
            x: { display: false },
            y: { display: false },
          },
    plugins: {
      datalabels: {
        display: false,
      },
      // TODO font size of labels - should be consistent
      legend: {
        // TODO display: false for single series
        display: type === "donut" || type === "pie" || data.datasets.length > 1,
        position: type === "donut" || type === "pie" ? "right" : "top",
        align: "center",
        labels: {
          boxWidth: 20,
          color: foreground,
          font: {
            size: 11,
          },
        },
      },
    },
  };
};

// const buildChartOptions = (
//   type,
//   properties: ChartProperties,
//   data,
//   min,
//   max,
//   theme,
//   themeWrapperRef
// ) => {
//   const baseOptions = getBaseOptions(
//     type,
//     data,
//     min,
//     max,
//     theme,
//     themeWrapperRef
//   );
//
//   return baseOptions;

// let overrideOptions = {};
//
// // Set legend options
// if (properties.legend) {
//   // Legend display setting
//   const legendDisplay = properties.legend.display;
//   if (legendDisplay === "always") {
//     overrideOptions = set(overrideOptions, "plugins.legend.display", true);
//   } else if (legendDisplay === "none") {
//     overrideOptions = set(overrideOptions, "plugins.legend.display", false);
//   }
//
//   // Legend display position
//   const legendPosition = properties.legend.position;
//   if (legendPosition === "top") {
//     overrideOptions = set(overrideOptions, "plugins.legend.position", "top");
//   } else if (legendPosition === "right") {
//     overrideOptions = set(
//       overrideOptions,
//       "plugins.legend.position",
//       "right"
//     );
//   } else if (legendPosition === "bottom") {
//     overrideOptions = set(
//       overrideOptions,
//       "plugins.legend.position",
//       "bottom"
//     );
//   } else if (legendPosition === "left") {
//     overrideOptions = set(overrideOptions, "plugins.legend.position", "left");
//   }
// }
//
// // Axes settings
// if (properties.axes) {
//   // X axis settings
//   if (properties.axes.x) {
//     // X axis display setting
//     if (properties.axes.x.display) {
//       const xAxisDisplay = properties.axes.x.display;
//       if (xAxisDisplay === "always") {
//         overrideOptions = set(overrideOptions, "scales.x.display", "always");
//       } else if (xAxisDisplay === "none") {
//         overrideOptions = set(overrideOptions, "scales.x.display", "none");
//       }
//     }
//
//     // X axis title settings
//     if (properties.axes.x.title) {
//       // X axis title display setting
//       const xAxisTitleDisplay = properties.axes.x.title.display;
//       if (xAxisTitleDisplay === "always") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.x.title.display",
//           true
//         );
//       } else if (xAxisTitleDisplay === "none") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.x.title.display",
//           false
//         );
//       }
//
//       // X Axis title align setting
//       const xAxisTitleAlign = properties.axes.x.title.align;
//       if (xAxisTitleAlign === "start") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.x.title.align",
//           "start"
//         );
//       } else if (xAxisTitleAlign === "center") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.x.title.align",
//           "center"
//         );
//       } else if (xAxisTitleAlign === "end") {
//         overrideOptions = set(overrideOptions, "scales.x.title.align", "end");
//       }
//
//       // X Axis title value setting
//       const xAxisTitleValue = properties.axes.x.title.value;
//       if (xAxisTitleValue) {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.x.title.text",
//           xAxisTitleValue
//         );
//       }
//     }
//
//     // X axis labels settings
//     if (properties.axes.x.labels) {
//       // X axis labels display setting
//       const xAxisTicksDisplay = properties.axes.x.labels.display;
//       if (xAxisTicksDisplay === "always") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.x.ticks.display",
//           true
//         );
//       } else if (xAxisTicksDisplay === "none") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.x.ticks.display",
//           false
//         );
//       }
//     }
//   }
//
//   // Y axis settings
//   if (properties.axes.y) {
//     // Y axis display setting
//     if (properties.axes.y.display) {
//       const yAxisDisplay = properties.axes.y.display;
//       if (yAxisDisplay === "always") {
//         overrideOptions = set(overrideOptions, "scales.y.display", "always");
//       } else if (yAxisDisplay === "none") {
//         overrideOptions = set(overrideOptions, "scales.y.display", "none");
//       }
//     }
//
//     // Y axis title settings
//     if (properties.axes.y.title) {
//       // Y axis title display setting
//       const yAxisTitleDisplay = properties.axes.y.title.display;
//       if (yAxisTitleDisplay === "always") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.y.title.display",
//           true
//         );
//       } else if (yAxisTitleDisplay === "none") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.y.title.display",
//           false
//         );
//       }
//
//       // Y Axis title align setting
//       const yAxisTitleAlign = properties.axes.y.title.align;
//       if (yAxisTitleAlign === "start") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.y.title.align",
//           "start"
//         );
//       } else if (yAxisTitleAlign === "center") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.y.title.align",
//           "center"
//         );
//       } else if (yAxisTitleAlign === "end") {
//         overrideOptions = set(overrideOptions, "scales.y.title.align", "end");
//       }
//
//       // Y Axis title value setting
//       const yAxisTitleValue = properties.axes.y.title.value;
//       if (yAxisTitleValue) {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.y.title.text",
//           yAxisTitleValue
//         );
//       }
//     }
//
//     // Y axis labels settings
//     if (properties.axes.y.labels) {
//       // Y axis labels display setting
//       const yAxisTicksDisplay = properties.axes.y.labels.display;
//       if (yAxisTicksDisplay === "always") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.y.ticks.display",
//           true
//         );
//       } else if (yAxisTicksDisplay === "none") {
//         overrideOptions = set(
//           overrideOptions,
//           "scales.y.ticks.display",
//           false
//         );
//       }
//     }
//
//     // Y Axis min value setting
//     if (has(properties, "axes.y.min")) {
//       overrideOptions = set(
//         overrideOptions,
//         "scales.y.min",
//         get(properties, "axes.y.min")
//       );
//     }
//
//     // Y Axis max value setting
//     if (has(properties, "axes.y.max")) {
//       overrideOptions = set(
//         overrideOptions,
//         "scales.y.max",
//         get(properties, "axes.y.max")
//       );
//     }
//   }
// }
//
// // Grouping setting
// if (properties.grouping) {
//   const groupingSetting = properties.grouping;
//   if (groupingSetting === "compare") {
//     overrideOptions = set(overrideOptions, "scales.x.stacked", false);
//     overrideOptions = set(overrideOptions, "scales.y.stacked", false);
//   }
// }
//
// // return { scales: { y: { min, max } } };
// return merge(baseOptions, overrideOptions);
// };

// const buildChartInputs = (
//   rawData: LeafNodeData,
//   inputs: ChartProperties,
//   theme,
//   themeWrapperRef
// ) => {
//   return {
//     animation: false,
//     title: {
//       text: "Referer of a Website",
//       subtext: "Fake Data",
//       left: "center",
//     },
//     tooltip: {
//       trigger: "item",
//     },
//     legend: {
//       orient: "vertical",
//       left: "left",
//     },
//     series: [
//       {
//         name: "Access From",
//         type: "pie",
//         radius: "50%",
//         data: [
//           { value: 1048, name: "Search Engine" },
//           { value: 735, name: "Direct" },
//           { value: 580, name: "Email" },
//           { value: 484, name: "Union Ads" },
//           { value: 300, name: "Video Ads" },
//         ],
//         emphasis: {
//           itemStyle: {
//             shadowBlur: 10,
//             shadowOffsetX: 0,
//             shadowColor: "rgba(0, 0, 0, 0.5)",
//           },
//         },
//       },
//     ],
//   };
//
//   if (!rawData) {
//     return null;
//   }
//   const {
//     min,
//     max,
//     options: dataOptions,
//   } = buildChartDataInputs(rawData, inputs.type, inputs);
//   const options = buildChartOptions(
//     inputs.type,
//     inputs,
//     dataOptions,
//     min,
//     max,
//     theme,
//     themeWrapperRef
//   );
//   return {
//     ...options,
//     ...dataOptions,
//   };
//   // return options;
// };

// const toChartJSType = (type: ChartType): ChartJSType => {
//   // A column chart in chart.js is a bar chart with different options
//   if (type === "column") {
//     return "bar";
//   }
//   // Different spelling
//   if (type === "donut") {
//     return "doughnut";
//   }
//   return type as ChartJSType;
// };

interface ChartComponentProps {
  options: EChartsOption;
  theme: Theme;
  themeWrapperRef: React.Ref<null>;
}

const Chart = ({ options, theme, themeWrapperRef }: ChartComponentProps) => {
  const [_, setRandomVal] = useState(0);
  const chartRef = useRef<ReactEChartsCore>(null);
  const [imageUrl, setImageUrl] = useState<string | null>(null);
  const [showZoom, setShowZoom] = useState(false);
  const { definition: panelDefinition, showExpand } = usePanel();
  const { dispatch } = useReport();
  const mediaMode = useMediaMode();

  // const options = buildChartInputs(data, inputs, theme, themeWrapperRef);

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  useEffect(() => {
    if (!chartRef.current || !options) {
      return;
    }

    const echartInstance = chartRef.current.getEchartsInstance();
    setImageUrl(echartInstance.getDataURL());
  }, [chartRef, options]);

  if (!options) {
    return null;
  }

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
            className="chart-canvas"
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

const ChartWrapper = (props: ChartProps) => {
  const { theme, wrapperRef } = useTheme();

  if (!wrapperRef) {
    return null;
  }

  if (!props.data) {
    return null;
  }

  return (
    <Chart
      options={buildChartOptions(props, theme, wrapperRef)}
      theme={theme}
      themeWrapperRef={wrapperRef}
    />
  );
};

type ChartDefinition = PanelDefinition & {
  properties: ChartProperties;
};

const renderChart = (definition: ChartDefinition) => {
  // We default to column charts if not specified
  const {
    properties: { type = "column" },
  } = definition;
  const chart = Charts[type];

  if (!chart) {
    return <ErrorPanel error={`Unknown chart type ${type}`} />;
  }

  const Component = chart.component;
  return <Component {...definition} />;
};

const RenderChart = (props: ChartDefinition) => {
  return renderChart(props);
};

export default ChartWrapper;

export { buildChartOptions, RenderChart };
