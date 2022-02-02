import Charts, {
  ChartJSType,
  ChartProperties,
  ChartProps,
  ChartType,
} from "../index";
import ErrorPanel from "../../Error";
import useMediaMode from "../../../../hooks/useMediaMode";
import { Chart as ChartJS, registerables } from "chart.js";
import { buildChartDataInputs, LeafNodeData } from "../../common";
import { Chart as ReactChartJS } from "react-chartjs-2";
import { PanelDefinition, useReport } from "../../../../hooks/useReport";
import { get, has, merge, set } from "lodash";
import { useEffect, useRef, useState } from "react";
import { usePanel } from "../../../../hooks/usePanel";
import { useTheme } from "../../../../hooks/useTheme";
import { ZoomIcon } from "../../../../constants/icons";

ChartJS.register(...registerables);

ChartJS.defaults.font.size = 12.25;

const getBaseOptions = (type, data, min, max, theme, themeWrapperRef) => {
  // We need to get the theme CSS variable values - these are accessible on the theme root element and below in the tree
  // @ts-ignore
  const style = window.getComputedStyle(themeWrapperRef);
  const foreground = style.getPropertyValue("--color-foreground");
  const foregroundLightest = style.getPropertyValue(
    "--color-foreground-lightest"
  );
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

const buildChartOptions = (
  type,
  properties: ChartProperties,
  data,
  min,
  max,
  theme,
  themeWrapperRef
) => {
  const baseOptions = getBaseOptions(
    type,
    data,
    min,
    max,
    theme,
    themeWrapperRef
  );
  let overrideOptions = {};

  // Set legend options
  if (properties.legend) {
    // Legend display setting
    const legendDisplay = properties.legend.display;
    if (legendDisplay === "always") {
      overrideOptions = set(overrideOptions, "plugins.legend.display", true);
    } else if (legendDisplay === "none") {
      overrideOptions = set(overrideOptions, "plugins.legend.display", false);
    }

    // Legend display position
    const legendPosition = properties.legend.position;
    if (legendPosition === "top") {
      overrideOptions = set(overrideOptions, "plugins.legend.position", "top");
    } else if (legendPosition === "right") {
      overrideOptions = set(
        overrideOptions,
        "plugins.legend.position",
        "right"
      );
    } else if (legendPosition === "bottom") {
      overrideOptions = set(
        overrideOptions,
        "plugins.legend.position",
        "bottom"
      );
    } else if (legendPosition === "left") {
      overrideOptions = set(overrideOptions, "plugins.legend.position", "left");
    }
  }

  // Axes settings
  if (properties.axes) {
    // X axis settings
    if (properties.axes.x) {
      // X axis display setting
      if (properties.axes.x.display) {
        const xAxisDisplay = properties.axes.x.display;
        if (xAxisDisplay === "always") {
          overrideOptions = set(overrideOptions, "scales.x.display", "always");
        } else if (xAxisDisplay === "none") {
          overrideOptions = set(overrideOptions, "scales.x.display", "none");
        }
      }

      // X axis title settings
      if (properties.axes.x.title) {
        // X axis title display setting
        const xAxisTitleDisplay = properties.axes.x.title.display;
        if (xAxisTitleDisplay === "always") {
          overrideOptions = set(
            overrideOptions,
            "scales.x.title.display",
            true
          );
        } else if (xAxisTitleDisplay === "none") {
          overrideOptions = set(
            overrideOptions,
            "scales.x.title.display",
            false
          );
        }

        // X Axis title align setting
        const xAxisTitleAlign = properties.axes.x.title.align;
        if (xAxisTitleAlign === "start") {
          overrideOptions = set(
            overrideOptions,
            "scales.x.title.align",
            "start"
          );
        } else if (xAxisTitleAlign === "center") {
          overrideOptions = set(
            overrideOptions,
            "scales.x.title.align",
            "center"
          );
        } else if (xAxisTitleAlign === "end") {
          overrideOptions = set(overrideOptions, "scales.x.title.align", "end");
        }

        // X Axis title value setting
        const xAxisTitleValue = properties.axes.x.title.value;
        if (xAxisTitleValue) {
          overrideOptions = set(
            overrideOptions,
            "scales.x.title.text",
            xAxisTitleValue
          );
        }
      }

      // X axis labels settings
      if (properties.axes.x.labels) {
        // X axis labels display setting
        const xAxisTicksDisplay = properties.axes.x.labels.display;
        if (xAxisTicksDisplay === "always") {
          overrideOptions = set(
            overrideOptions,
            "scales.x.ticks.display",
            true
          );
        } else if (xAxisTicksDisplay === "none") {
          overrideOptions = set(
            overrideOptions,
            "scales.x.ticks.display",
            false
          );
        }
      }
    }

    // Y axis settings
    if (properties.axes.y) {
      // Y axis display setting
      if (properties.axes.y.display) {
        const yAxisDisplay = properties.axes.y.display;
        if (yAxisDisplay === "always") {
          overrideOptions = set(overrideOptions, "scales.y.display", "always");
        } else if (yAxisDisplay === "none") {
          overrideOptions = set(overrideOptions, "scales.y.display", "none");
        }
      }

      // Y axis title settings
      if (properties.axes.y.title) {
        // Y axis title display setting
        const yAxisTitleDisplay = properties.axes.y.title.display;
        if (yAxisTitleDisplay === "always") {
          overrideOptions = set(
            overrideOptions,
            "scales.y.title.display",
            true
          );
        } else if (yAxisTitleDisplay === "none") {
          overrideOptions = set(
            overrideOptions,
            "scales.y.title.display",
            false
          );
        }

        // Y Axis title align setting
        const yAxisTitleAlign = properties.axes.y.title.align;
        if (yAxisTitleAlign === "start") {
          overrideOptions = set(
            overrideOptions,
            "scales.y.title.align",
            "start"
          );
        } else if (yAxisTitleAlign === "center") {
          overrideOptions = set(
            overrideOptions,
            "scales.y.title.align",
            "center"
          );
        } else if (yAxisTitleAlign === "end") {
          overrideOptions = set(overrideOptions, "scales.y.title.align", "end");
        }

        // Y Axis title value setting
        const yAxisTitleValue = properties.axes.y.title.value;
        if (yAxisTitleValue) {
          overrideOptions = set(
            overrideOptions,
            "scales.y.title.text",
            yAxisTitleValue
          );
        }
      }

      // Y axis labels settings
      if (properties.axes.y.labels) {
        // Y axis labels display setting
        const yAxisTicksDisplay = properties.axes.y.labels.display;
        if (yAxisTicksDisplay === "always") {
          overrideOptions = set(
            overrideOptions,
            "scales.y.ticks.display",
            true
          );
        } else if (yAxisTicksDisplay === "none") {
          overrideOptions = set(
            overrideOptions,
            "scales.y.ticks.display",
            false
          );
        }
      }

      // Y Axis min value setting
      if (has(properties, "axes.y.min")) {
        overrideOptions = set(
          overrideOptions,
          "scales.y.min",
          get(properties, "axes.y.min")
        );
      }

      // Y Axis max value setting
      if (has(properties, "axes.y.max")) {
        overrideOptions = set(
          overrideOptions,
          "scales.y.max",
          get(properties, "axes.y.max")
        );
      }
    }
  }

  // Grouping setting
  if (properties.grouping) {
    const groupingSetting = properties.grouping;
    if (groupingSetting === "compare") {
      overrideOptions = set(overrideOptions, "scales.x.stacked", false);
      overrideOptions = set(overrideOptions, "scales.y.stacked", false);
    }
  }

  // return { scales: { y: { min, max } } };
  return merge(baseOptions, overrideOptions);
};

const buildChartInputs = (
  rawData: LeafNodeData,
  inputs: ChartProperties,
  theme,
  themeWrapperRef
) => {
  if (!rawData) {
    return null;
  }
  const { data, min, max } = buildChartDataInputs(rawData, inputs.type, inputs);
  const options = buildChartOptions(
    inputs.type,
    inputs,
    data,
    min,
    max,
    theme,
    themeWrapperRef
  );
  return { data, options };
};

const toChartJSType = (type: ChartType): ChartJSType => {
  // A column chart in chart.js is a bar chart with different options
  if (type === "column") {
    return "bar";
  }
  // Different spelling
  if (type === "donut") {
    return "doughnut";
  }
  return type as ChartJSType;
};

const Chart = ({ data, inputs, theme, themeWrapperRef }) => {
  const [_, setRandomVal] = useState(0);
  const chartRef = useRef<ChartJS>(null);
  const [imageUrl, setImageUrl] = useState<string | null>(null);
  const [showZoom, setShowZoom] = useState(false);
  const { definition: panelDefinition, showExpand } = usePanel();
  const { dispatch } = useReport();
  const mediaMode = useMediaMode();

  const built = buildChartInputs(data, inputs, theme, themeWrapperRef);

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  useEffect(() => {
    if (!chartRef.current || !built) {
      return;
    }

    setImageUrl(chartRef.current.toBase64Image());
  }, [chartRef, inputs, built]);

  if (!built) {
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
              className="absolute right-0 top-0 cursor-pointer"
              onClick={() =>
                dispatch({ type: "select_panel", panel: panelDefinition })
              }
            >
              <ZoomIcon className="h-5 w-5 text-black-scale-4" />
            </div>
          )}
          <ReactChartJS
            ref={chartRef}
            className="chart-canvas"
            type={toChartJSType(inputs.type)}
            data={built.data}
            // @ts-ignore
            options={built.options}
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

const ChartWrapper = ({ data, inputs }) => {
  const { theme, wrapperRef } = useTheme();

  if (!wrapperRef) {
    return null;
  }

  if (!data) {
    return null;
  }

  return (
    <Chart
      data={data}
      inputs={inputs}
      theme={theme}
      themeWrapperRef={wrapperRef}
    />
  );
};

type ChartDefinition = PanelDefinition & {
  properties: ChartProps;
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

export { RenderChart };
