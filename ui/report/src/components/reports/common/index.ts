import moment from "moment";
import { ChartProperties, ChartType } from "../charts";
import { ColorGenerator } from "../../../utils/color";
import { HierarchyProperties, HierarchyType } from "../hierarchies";
import { raw } from "@storybook/react";

export type Width = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12;

export interface BasePrimitiveProps {
  base?: string;
  name: string;
  node_type: string;
  title?: string;
  width?: Width;
}

export interface LeafNodeDataColumn {
  name: string;
  data_type_name: string;
}

export interface HierarchyDataRow {
  id: string;
  category: string;
  parent: string | null;
}

export interface HierarchyDataRowEdge {
  source: string;
  target: string;
  value: number;
}

export type LeafNodeDataRow = any[];

export interface LeafNodeData {
  columns: LeafNodeDataColumn[];
  rows: LeafNodeDataRow[];
}

export interface ExecutablePrimitiveProps {
  sql?: string;
  data?: LeafNodeData;
  error?: Error;
}

interface SeriesData {
  label: string;
  data: any[];
  backgroundColor: string | string[];
}

interface SeriesLookup {
  [series: string]: {
    value: any;
  };
}

interface SeriesTimeLookup {
  [time: string]: SeriesLookup;
}

interface TotalLookup {
  [key: string]: Scale;
}

interface Scale {
  min: number;
  max: number;
}

const buildSeriesInputs = (rawData, seriesDataFormat, seriesDataType) => {
  const seriesDataLookup: SeriesTimeLookup = {};
  const seriesDataTotalLookup: TotalLookup = {};
  const seriesLabels: string[] = [];
  const timeSeriesLabels: string[] = [];
  if (seriesDataFormat === "row") {
    if (seriesDataType === "time") {
      for (const row of rawData.slice(1)) {
        const timeRaw = row[0];
        const formattedTime = moment(timeRaw).format("DD MMM YYYY");
        const series = row[1];
        const value = row[2];
        seriesDataLookup[formattedTime] = seriesDataLookup[formattedTime] || {};
        const timeEntry = seriesDataLookup[formattedTime];
        timeEntry[series] = timeEntry[series] || {};
        const seriesEntry = timeEntry[series];
        seriesEntry.value = value;
        if (timeSeriesLabels.indexOf(formattedTime) === -1) {
          timeSeriesLabels.push(formattedTime);
        }
        if (value !== 0 && seriesLabels.indexOf(series) === -1) {
          seriesLabels.push(series);
        }

        seriesDataTotalLookup[formattedTime] = seriesDataTotalLookup[
          formattedTime
        ] || {
          min: 0,
          max: 0,
        };
        if (value > 0) {
          seriesDataTotalLookup[formattedTime].max += value;
        }
        if (value < 0) {
          seriesDataTotalLookup[formattedTime].min += value;
        }
      }
    }
  }

  return {
    seriesDataLookup,
    seriesDataTotalLookup,
    seriesLabels,
    timeSeriesLabels,
  };
};

const buildSeriesDataInputs = (rawData, seriesDataFormat, seriesDataType) => {
  const {
    seriesDataLookup,
    seriesDataTotalLookup,
    seriesLabels,
    timeSeriesLabels,
  } = buildSeriesInputs(rawData, seriesDataFormat, seriesDataType);
  const datasets: SeriesData[] = [];

  for (
    let seriesLabelIndex = 0;
    seriesLabelIndex < seriesLabels.length;
    seriesLabelIndex++
  ) {
    const seriesLabel = seriesLabels[seriesLabelIndex];
    const data: any[] = [];

    for (const timeSeriesLabel of timeSeriesLabels) {
      const timeSeriesEntry = seriesDataLookup[timeSeriesLabel];
      const seriesEntry = timeSeriesEntry[seriesLabel];
      data.push(seriesEntry ? seriesEntry.value : null);
    }

    datasets.push({
      label: seriesLabel,
      data,
      backgroundColor: themeColors[seriesLabelIndex],
    });
  }

  let min = 0;
  let max = 0;
  for (const {
    min: seriesDataMinTotal,
    max: seriesDataMaxTotal,
  } of Object.values(seriesDataTotalLookup)) {
    if (seriesDataMinTotal < min) {
      min = seriesDataMinTotal;
    }
    if (seriesDataMaxTotal > max) {
      max = seriesDataMaxTotal;
    }
  }

  min = adjustMinValue(min);
  max = adjustMaxValue(max);

  const data = {
    labels: timeSeriesLabels,
    datasets,
  };

  return {
    data,
    min,
    max,
  };
};

const adjust = (value, divisor, direction = "asc") => {
  const remainder = value % divisor;
  if (direction === "asc") {
    return remainder === 0 ? value + divisor : value + (divisor - remainder);
  } else {
    return remainder === 0 ? value - divisor : value - (divisor + remainder);
  }
};

const adjustMinValue = (initial) => {
  if (initial >= 0) {
    return 0;
  }

  let min = initial;
  if (initial <= -10000) {
    min = adjust(min, 1000, "desc");
  } else if (initial <= -1000) {
    min = adjust(min, 100, "desc");
  } else if (initial <= -200) {
    min = adjust(min, 50, "desc");
  } else if (initial <= -50) {
    min = adjust(min, 10, "desc");
  } else if (initial <= -20) {
    min = adjust(min, 5, "desc");
  } else if (initial <= -10) {
    min = adjust(min, 2, "desc");
  } else {
    min -= 1;
  }
  return min;
};

const adjustMaxValue = (initial) => {
  if (initial <= 0) {
    return 0;
  }

  let max = initial;
  if (initial < 10) {
    max += 1;
  } else if (initial < 20) {
    max = adjust(max, 2);
  } else if (initial < 50) {
    max = adjust(max, 5);
  } else if (initial < 200) {
    max = adjust(max, 10);
  } else if (initial < 1000) {
    max = adjust(max, 50);
  } else if (initial < 10000) {
    max = adjust(max, 100);
  } else {
    max = adjust(max, 1000);
  }
  return max;
};

const buildChartDataInputs = (
  rawData: LeafNodeData,
  type: ChartType,
  properties: ChartProperties
) => {
  const seriesLength = rawData.columns.length - 1;

  const labels: string[] = [];
  const datasets: SeriesData[] = [];

  let min = 0;
  let max = 0;

  for (const row of rawData.rows) {
    labels.push(row[0]);
  }
  for (let seriesIndex = 1; seriesIndex <= seriesLength; seriesIndex++) {
    const data: any[] = [];
    for (const row of rawData.rows) {
      const dataValue = row[seriesIndex];
      if (dataValue < min) {
        min = dataValue;
      }
      if (dataValue > max) {
        max = dataValue;
      }
      data.push(dataValue);
    }
    const seriesName = rawData.columns[seriesIndex].name;
    let seriesOverrides;
    if (properties.series && properties.series[seriesName]) {
      seriesOverrides = properties.series[seriesName];
    }
    datasets.push({
      label:
        seriesOverrides && seriesOverrides.title
          ? seriesOverrides.title
          : seriesName,
      data,
      backgroundColor:
        type === "donut" || type === "pie"
          ? themeColors.slice(0, labels.length)
          : seriesOverrides && seriesOverrides.color
          ? seriesOverrides.color
          : themeColors[seriesIndex - 1],
    });
  }
  const data = {
    labels,
    datasets,
  };

  // Adjust min and max to allow breathing space in the chart
  min = adjustMinValue(min);
  max = adjustMaxValue(max);

  return {
    data,
    min,
    max,
  };
};

const buildHierarchyDataInputs = (
  rawData: LeafNodeData,
  type: HierarchyType,
  properties: HierarchyProperties
) => {
  let colorIndex = 0;
  const builtData = [];
  const categories = {};
  const usedIds = {};
  const objectData = rawData.rows.map((dataRow) => {
    const row: HierarchyDataRow = { parent: null, category: "", id: "" };
    for (let colIndex = 0; colIndex < rawData.columns.length; colIndex++) {
      const column = rawData.columns[colIndex];
      row[column.name] = dataRow[colIndex];
    }

    if (row.category && !categories[row.category]) {
      categories[row.category] = { color: themeColors[colorIndex++] };
    }

    if (!usedIds[row.id]) {
      builtData.push({
        ...row,
        itemStyle: {
          // @ts-ignore
          color: categories[row.category].color,
        },
      });
      usedIds[row.id] = true;
    }
    return row;
  });
  const edges: HierarchyDataRowEdge[] = [];
  const edgeValues = {};
  for (const d of objectData) {
    // TODO remove <null> after Kai fixes base64 issue and removes col string conversion
    if (d.parent === null || d.parent === "<null>") {
      d.parent = null;
      continue;
    }
    edges.push({ source: d.parent, target: d.id, value: 0.01 });
    edgeValues[d.parent] = (edgeValues[d.parent] || 0) + 0.01;
  }
  for (const e of edges) {
    var v = 0;
    if (edgeValues[e.target]) {
      for (const e2 of edges) {
        if (e.target === e2.source) {
          v += edgeValues[e2.target] || 0.01;
        }
      }
      e.value = v;
    }
  }

  return {
    data: builtData,
    links: edges,
  };
};

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

export {
  adjustMinValue,
  adjustMaxValue,
  buildChartDataInputs,
  buildHierarchyDataInputs,
  buildSeriesDataInputs,
  buildSeriesInputs,
  themeColors,
};
