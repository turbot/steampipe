import moment from "moment";
import { ChartProperties, ChartType } from "../charts";
import { themeColors } from "../charts/Chart";

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

export interface LeafNodeDataItem {
  [key: string]: any;
}

export interface LeafNodeData {
  columns: LeafNodeDataColumn[];
  items: LeafNodeDataItem[];
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

  const firstCol = rawData.columns[0];

  for (const row of rawData.items) {
    labels.push(row[firstCol.name]);
  }
  for (let seriesIndex = 1; seriesIndex <= seriesLength; seriesIndex++) {
    const data: any[] = [];
    const colForIndex = rawData.columns[seriesIndex];
    for (const row of rawData.items) {
      const dataValue = row[colForIndex.name];
      if (dataValue < min) {
        min = dataValue;
      }
      if (dataValue > max) {
        max = dataValue;
      }
      data.push(dataValue);
    }
    const seriesName = colForIndex.name;
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

export {
  adjustMinValue,
  adjustMaxValue,
  buildChartDataInputs,
  buildSeriesDataInputs,
  buildSeriesInputs,
};
