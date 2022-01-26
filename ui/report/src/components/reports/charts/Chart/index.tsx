import Charts, {
  ChartJSType,
  ChartProperties,
  ChartProps,
  ChartType,
} from "../index";
import ErrorPanel from "../../Error";
import Icon from "../../../Icon";
import useMediaMode from "../../../../hooks/useMediaMode";
import {
  ArcElement,
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Title,
  Tooltip,
} from "chart.js";
import { buildChartDataInputs, LeafNodeData } from "../../common";
import { Chart as ReactChartJS } from "react-chartjs-2";
import { ColorGenerator } from "../../../../utils/color";
import { get, has, merge, property, set } from "lodash";
import { useEffect, useRef, useState } from "react";
import { usePanel } from "../../../../hooks/usePanel";
import { PanelDefinition, useReport } from "../../../../hooks/useReport";
import { useTheme } from "../../../../hooks/useTheme";
import { zoomIcon } from "../../../../constants/icons";

ChartJS.register(
  ArcElement,
  BarElement,
  CategoryScale,
  // ChartDataLabels,
  LineElement,
  LinearScale,
  PointElement,
  Title,
  Tooltip,
  Legend
);

ChartJS.defaults.font.size = 12.25;

// TODO color scheme - need to find something better?
const generateColors = () => {
  // return [
  //   "#6388b4",
  //   "#ffae34",
  //   "#ef6f6a",
  //   "#8cc2ca",
  //   "#55ad89",
  //   "#c3bc3f",
  //   "#bb7693",
  //   "#baa094",
  //   "#a9b5ae",
  //   "#767676",
  // ];
  // return [
  //   "#4f6980",
  //   "#849db1",
  //   "#a2ceaa",
  //   "#638b66",
  //   "#bfbb60",
  //   "#f47942",
  //   "#fbb04e",
  //   "#b66353",
  //   "#d7ce9f",
  //   "#b9aa97",
  //   "#7e756d",
  // ];
  // return [
  //   "#1f77b4",
  //   "#aec7e8",
  //   "#ff7f0e",
  //   "#ffbb78",
  //   "#2ca02c",
  //   "#98df8a",
  //   "#d62728",
  //   "#ff9896",
  //   "#9467bd",
  //   "#c5b0d5",
  //   "#8c564b",
  //   "#c49c94",
  //   "#e377c2",
  //   "#f7b6d2",
  //   "#7f7f7f",
  //   "#c7c7c7",
  //   "#bcbd22",
  //   "#dbdb8d",
  //   "#17becf",
  //   "#9edae5",
  // ];
  // tableau.Tableau20
  return [
    "#4E79A7",
    "#A0CBE8",
    "#F28E2B",
    "#FFBE7D",
    "#59A14F",
    "#8CD17D",
    "#B6992D",
    "#F1CE63",
    "#499894",
    "#86BCB6",
    "#E15759",
    "#FF9D9A",
    "#79706E",
    "#BAB0AC",
    "#D37295",
    "#FABFD2",
    "#B07AA1",
    "#D4A6C8",
    "#9D7660",
    "#D7B5A6",
  ];
  const colorGenerator = new ColorGenerator(24, 5);
  const colors: string[] = [];
  for (let i = 0; i < 20; i++) {
    const nextColor = colorGenerator.nextColor();
    colors.push(nextColor.hex);
  }
  return colors;
};

const themeColors = generateColors();

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

const buildInputs = (
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
  const { definition: panelDefinition, showExpand } = usePanel();
  const { dispatch } = useReport();
  const mediaMode = useMediaMode();

  const built = buildInputs(data, inputs, theme, themeWrapperRef);

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  useEffect(() => {
    if (!chartRef.current || !built) {
      return;
    }

    setImageUrl(chartRef.current.toBase64Image());
  }, [chartRef, inputs, built]);

  const [showZoom, setShowZoom] = useState(false);

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
              <Icon icon={zoomIcon} />
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

export type ChartDefinition = PanelDefinition & {
  properties: ChartProps;
};

const renderChart = (definition: ChartDefinition) => {
  const {
    properties: { type },
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

export { RenderChart, themeColors };
