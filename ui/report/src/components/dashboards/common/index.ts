import { ChartProperties, ChartTransform, ChartType } from "../charts";
import { HierarchyProperties, HierarchyType } from "../hierarchies";
import { sortBy } from "lodash";

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
  name: string | null;
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

export type EChartsType = "bar" | "line" | "pie" | "sankey";

const toEChartsType = (type: ChartType | HierarchyType): EChartsType => {
  // A column chart in chart.js is a bar chart with different options
  if (type === "column") {
    return "bar";
  }
  // Different spelling
  if (type === "donut") {
    return "pie";
  }
  return type as EChartsType;
};

interface ChartDatasetResponse {
  dataset: any[][];
  rowSeriesLabels: string[];
  transform: ChartTransform;
}

const crosstabDataTransform = (data: LeafNodeData): ChartDatasetResponse => {
  if (data.columns.length < 3) {
    return { dataset: [], rowSeriesLabels: [], transform: "none" };
  }
  const xAxis = {};
  const series = {};
  const xAxisLabels: string[] = [];
  const seriesLabels: string[] = [];
  for (const row of data.rows) {
    const xAxisLabel = row[0];
    const seriesName = row[1];
    const seriesValue = row[2];

    if (!xAxis[xAxisLabel]) {
      xAxis[xAxisLabel] = {};
      xAxisLabels.push(xAxisLabel);
    }

    xAxis[xAxisLabel] = xAxis[xAxisLabel] || {};

    if (seriesName) {
      const existing = xAxis[xAxisLabel][seriesName];
      xAxis[xAxisLabel][seriesName] = existing
        ? existing + seriesValue
        : seriesValue;

      if (!series[seriesName]) {
        series[seriesName] = true;
        seriesLabels.push(seriesName);
      }
    }
  }

  const dataset: any[] = [];
  const headerRow: string[] = [];
  headerRow.push(data.columns[0].name);
  for (const seriesLabel of seriesLabels) {
    headerRow.push(seriesLabel);
  }
  dataset.push(headerRow);

  for (const xAxisLabel of xAxisLabels) {
    const row = [xAxisLabel];
    for (const seriesLabel of seriesLabels) {
      const seriesValue = xAxis[xAxisLabel][seriesLabel];
      row.push(seriesValue === undefined ? null : seriesValue);
    }
    dataset.push(row);
  }

  return { dataset, rowSeriesLabels: seriesLabels, transform: "crosstab" };
};

const defaultDataTransform = (data: LeafNodeData): ChartDatasetResponse => {
  return {
    dataset: [data.columns.map((col) => col.name), ...data.rows],
    rowSeriesLabels: [],
    transform: "none",
  };
};

const isNumericCol = (data_type_name: string) => {
  return (
    data_type_name.toLowerCase().indexOf("int") >= 0 ||
    data_type_name.toLowerCase().indexOf("float") >= 0 ||
    data_type_name.toLowerCase().indexOf("numeric") >= 0
  );
};

const automaticDataTransform = (data: LeafNodeData): ChartDatasetResponse => {
  // We want to check if the data looks like something that can be crosstab transformed.
  // If that's 3 columns, with the first 2 non-numeric and the last numeric, we'll apply
  // a crosstab transform, else we'll apply the default transform
  if (data.columns.length === 3) {
    const col1Type = data.columns[0].data_type_name;
    const col2Type = data.columns[1].data_type_name;
    const col3Type = data.columns[2].data_type_name;
    if (
      !isNumericCol(col1Type) &&
      !isNumericCol(col2Type) &&
      isNumericCol(col3Type)
    ) {
      return crosstabDataTransform(data);
    }
  }
  return defaultDataTransform(data);
};

const buildChartDataset = (
  data: LeafNodeData | undefined,
  properties: ChartProperties | undefined
): ChartDatasetResponse => {
  if (!data || !data.columns) {
    return { dataset: [], rowSeriesLabels: [], transform: "none" };
  }

  const transform = properties?.transform;

  switch (transform) {
    case "crosstab":
      return crosstabDataTransform(data);
    case "none":
      return defaultDataTransform(data);
    // Must be not specified or "auto", which should check to see
    // if the data matches crosstab format and transform if it is
    default:
      return automaticDataTransform(data);
  }
};

// const buildSeriesInputs = (rawData, seriesDataFormat, seriesDataType) => {
//   const seriesDataLookup: SeriesTimeLookup = {};
//   const seriesDataTotalLookup: TotalLookup = {};
//   const seriesLabels: string[] = [];
//   const timeSeriesLabels: string[] = [];
//   if (seriesDataFormat === "row") {
//     if (seriesDataType === "time") {
//       for (const row of rawData.slice(1)) {
//         const timeRaw = row[0];
//         const formattedTime = moment(timeRaw).format("DD MMM YYYY");
//         const series = row[1];
//         const value = row[2];
//         seriesDataLookup[formattedTime] = seriesDataLookup[formattedTime] || {};
//         const timeEntry = seriesDataLookup[formattedTime];
//         timeEntry[series] = timeEntry[series] || {};
//         const seriesEntry = timeEntry[series];
//         seriesEntry.value = value;
//         if (timeSeriesLabels.indexOf(formattedTime) === -1) {
//           timeSeriesLabels.push(formattedTime);
//         }
//         if (value !== 0 && seriesLabels.indexOf(series) === -1) {
//           seriesLabels.push(series);
//         }
//
//         seriesDataTotalLookup[formattedTime] = seriesDataTotalLookup[
//           formattedTime
//         ] || {
//           min: 0,
//           max: 0,
//         };
//         if (value > 0) {
//           seriesDataTotalLookup[formattedTime].max += value;
//         }
//         if (value < 0) {
//           seriesDataTotalLookup[formattedTime].min += value;
//         }
//       }
//     }
//   }
//
//   return {
//     seriesDataLookup,
//     seriesDataTotalLookup,
//     seriesLabels,
//     timeSeriesLabels,
//   };
// };

// const buildSeriesDataInputs = (rawData, seriesDataFormat, seriesDataType) => {
//   const {
//     seriesDataLookup,
//     seriesDataTotalLookup,
//     seriesLabels,
//     timeSeriesLabels,
//   } = buildSeriesInputs(rawData, seriesDataFormat, seriesDataType);
//   const datasets: SeriesData[] = [];
//
//   for (
//     let seriesLabelIndex = 0;
//     seriesLabelIndex < seriesLabels.length;
//     seriesLabelIndex++
//   ) {
//     const seriesLabel = seriesLabels[seriesLabelIndex];
//     const data: any[] = [];
//
//     for (const timeSeriesLabel of timeSeriesLabels) {
//       const timeSeriesEntry = seriesDataLookup[timeSeriesLabel];
//       const seriesEntry = timeSeriesEntry[seriesLabel];
//       data.push(seriesEntry ? seriesEntry.value : null);
//     }
//
//     datasets.push({
//       name: seriesLabel,
//       data,
//       // backgroundColor: themeColors[seriesLabelIndex],
//     });
//   }
//
//   let min = 0;
//   let max = 0;
//   for (const {
//     min: seriesDataMinTotal,
//     max: seriesDataMaxTotal,
//   } of Object.values(seriesDataTotalLookup)) {
//     if (seriesDataMinTotal < min) {
//       min = seriesDataMinTotal;
//     }
//     if (seriesDataMaxTotal > max) {
//       max = seriesDataMaxTotal;
//     }
//   }
//
//   min = adjustMinValue(min);
//   max = adjustMaxValue(max);
//
//   const data = {
//     labels: timeSeriesLabels,
//     datasets,
//   };
//
//   return {
//     data,
//     min,
//     max,
//   };
// };

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

const buildSankeyDataInputs = (
  rawData: LeafNodeData,
  properties: HierarchyProperties | undefined
) => {
  let colorIndex = 0;
  const builtData = [];
  const categories = {};
  const usedIds = {};
  const objectData = rawData.rows.map((dataRow) => {
    const row: HierarchyDataRow = {
      parent: null,
      category: "",
      id: "",
      name: "",
    };
    for (let colIndex = 0; colIndex < rawData.columns.length; colIndex++) {
      const column = rawData.columns[colIndex];
      row[column.name] = dataRow[colIndex];
    }

    if (row.category && !categories[row.category]) {
      let color;
      if (
        properties &&
        properties.categories &&
        properties.categories[row.category] &&
        properties.categories[row.category].color
      ) {
        color = properties.categories[row.category].color;
        colorIndex++;
      } else {
        color = themeColors[colorIndex++];
      }
      categories[row.category] = { color };
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

interface Item {
  [key: string]: any;
}

interface TreeItem {
  [key: string]: Item | TreeItem[] | any;
}

// Taken from https://github.com/philipstanislaus/performant-array-to-tree
const arrayToTree = (rawData: LeafNodeData): TreeItem[] => {
  const rootParentIds = { "": true };

  // the resulting unflattened tree
  const rootItems: TreeItem[] = [];

  // stores all already processed items with their ids as key so we can easily look them up
  const lookup: { [id: string]: TreeItem } = {};

  // stores all item ids that have not been added to the resulting unflattened tree yet
  // this is an opt-in property, since it has a slight runtime overhead
  const orphanIds: null | Set<string | number> = new Set();

  // idea of this loop:
  // whenever an item has a parent, but the parent is not yet in the lookup object, we store a preliminary parent
  // in the lookup object and fill it with the data of the parent later
  // if an item has no parentId, add it as a root element to rootItems
  for (const rawRow of rawData.rows) {
    const row: HierarchyDataRow = {
      parent: null,
      category: "",
      id: "",
      name: "",
    };
    for (let colIndex = 0; colIndex < rawData.columns.length; colIndex++) {
      const column = rawData.columns[colIndex];
      row[column.name] = rawRow[colIndex];
    }

    const itemId = row.id;
    const parentId = row.parent;

    if (rootParentIds[itemId]) {
      throw new Error(
        `The row contains a node whose parent both exists in another node and is in ` +
          `\`rootParentIds\` (\`itemId\`: "${itemId}", \`rootParentIds\`: ${Object.keys(
            rootParentIds
          )
            .map((r) => `"${r}"`)
            .join(", ")}).`
      );
    }

    // look whether item already exists in the lookup table
    if (!Object.prototype.hasOwnProperty.call(lookup, itemId)) {
      // item is not yet there, so add a preliminary item (its data will be added later)
      lookup[itemId] = { children: [] };
    }

    // if we track orphans, delete this item from the orphan set if it is in it
    if (orphanIds) {
      orphanIds.delete(itemId);
    }

    // add the current item's data to the item in the lookup table

    lookup[itemId] = {
      ...row,
      children: lookup[itemId].children,
    };

    const treeItem = lookup[itemId];

    if (
      parentId === null ||
      parentId === undefined ||
      rootParentIds[parentId]
    ) {
      // is a root item
      rootItems.push(treeItem);
    } else {
      // has a parent

      // look whether the parent already exists in the lookup table
      if (!Object.prototype.hasOwnProperty.call(lookup, parentId)) {
        // parent is not yet there, so add a preliminary parent (its data will be added later)
        lookup[parentId] = { children: [] };

        // if we track orphans, add the generated parent to the orphan list
        if (orphanIds) {
          orphanIds.add(parentId);
        }
      }

      // add the current item to the parent
      lookup[parentId].children.push(treeItem);
    }
  }

  if (orphanIds?.size) {
    throw new Error(
      `The items array contains orphans that point to the following parentIds: ` +
        `[${Array.from(
          orphanIds
        )}]. These parentIds do not exist in the items array. Hint: prevent orphans to result ` +
        `in an error by passing the following option: { throwIfOrphans: false }`
    );
  }

  return rootItems;
};

const buildTreeDataInputs = (
  rawData: LeafNodeData,
  properties: HierarchyProperties | undefined
) => {
  const builtData: any = {};
  const tree = arrayToTree(rawData);
  return {
    data: tree,
  };
};

// TODO color scheme - need to find something better?
const generateColors = () => {
  // echarts vintage
  // return [
  //   "#d87c7c",
  //   "#919e8b",
  //   "#d7ab82",
  //   "#6e7074",
  //   "#61a0a8",
  //   "#efa18d",
  //   "#787464",
  //   "#cc7e63",
  //   "#724e58",
  //   "#4b565b",
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
};

const themeColors = generateColors();

export {
  adjustMinValue,
  adjustMaxValue,
  buildChartDataset,
  buildSankeyDataInputs,
  buildTreeDataInputs,
  themeColors,
  toEChartsType,
};
