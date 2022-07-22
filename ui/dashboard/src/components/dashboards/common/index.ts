import has from "lodash/has";
import { ChartProperties, ChartTransform, ChartType } from "../charts/types";
import { DashboardRunState } from "../../../hooks/useDashboard";
import { FlowProperties, FlowType } from "../flows/types";
import { getColumn } from "../../../utils/data";
import { GraphProperties, GraphType } from "../graphs/types";
import { HierarchyProperties, HierarchyType } from "../hierarchies/types";

export type Width = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12;

export interface BasePrimitiveProps {
  base?: string;
  name: string;
  panel_type: string;
  display_type?: string;
  title?: string;
  width?: Width;
}

export interface LeafNodeDataColumn {
  name: string;
  data_type: string;
}

export interface LeafNodeDataRow {
  [key: string]: any;
}

export interface LeafNodeData {
  columns: LeafNodeDataColumn[];
  rows: LeafNodeDataRow[];
}

export interface ExecutablePrimitiveProps {
  sql?: string;
  data?: LeafNodeData;
  error?: Error;
  status: DashboardRunState;
}

export type ColorOverride = "alert" | "info" | "ok" | string;

export type EChartsType = "bar" | "line" | "pie" | "sankey" | "tree" | "graph";

const toEChartsType = (
  type: ChartType | FlowType | GraphType | HierarchyType
): EChartsType => {
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
    const xAxisLabel = row[data.columns[0].name];
    const seriesName = row[data.columns[1].name];
    const seriesValue = row[data.columns[2].name];

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
    dataset: [
      data.columns.map((col) => col.name),
      ...data.rows.map((row) => data.columns.map((col) => row[col.name])),
    ],
    rowSeriesLabels: [],
    transform: "none",
  };
};

const isNumericCol = (data_type: string | null | undefined) => {
  if (!data_type) {
    return false;
  }
  return (
    data_type.toLowerCase().indexOf("int") >= 0 ||
    data_type.toLowerCase().indexOf("float") >= 0 ||
    data_type.toLowerCase().indexOf("numeric") >= 0
  );
};

const automaticDataTransform = (data: LeafNodeData): ChartDatasetResponse => {
  // We want to check if the data looks like something that can be crosstab transformed.
  // If that's 3 columns, with the first 2 non-numeric and the last numeric, we'll apply
  // a crosstab transform, else we'll apply the default transform
  if (data.columns.length === 3) {
    const col1Type = data.columns[0].data_type;
    const col2Type = data.columns[1].data_type;
    const col3Type = data.columns[2].data_type;
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

interface Node {
  id: string;
  title: string | null;
  category: string | null;
  depth: number | null;
  metadata: NodeMetadata | null;
  symbol: string | null;
}

interface Edge {
  id: string;
  from_id: string;
  to_id: string;
  title: string | null;
  category: string | null;
}

interface NodeMap {
  [id: string]: Node;
}

interface NodeMetadata {
  [key: string]: string;
}

interface EdgeMap {
  [edge_id: string]: boolean;
}

interface Category {
  color: string | null;
  icon: string | null;
}

interface CategoryMap {
  [category: string]: Category;
}

interface NodesAndEdgesMetadata {
  has_multiple_roots: boolean;
  contains_duplicate_edges: boolean;
}

export interface NodesAndEdges {
  nodes: Node[];
  root_nodes: NodeMap;
  edges: Edge[];
  categories: CategoryMap;
  metadata?: NodesAndEdgesMetadata;
  next_color_index: number;
}

const recordEdge = (
  edge_lookup,
  from_id: string,
  to_id: string,
  title: string | null = null,
  category: string | null = null
) => {
  let duplicate_edge = false;
  // Find any existing edge
  const edge_id = `${from_id}:${to_id}`;
  const existingNode = edge_lookup[edge_id];
  if (existingNode) {
    duplicate_edge = true;
  } else {
    edge_lookup[edge_id] = true;
  }

  const edge: Edge = {
    id: edge_id,
    from_id,
    to_id,
    title,
    category,
  };
  return {
    edge,
    duplicate_edge,
  };
};

const createNode = (
  id: string,
  title: string | null = null,
  category: string | null = null,
  depth: number | null = null,
  metadata: NodeMetadata | null = {},
  categories: CategoryMap = {}
) => {
  let symbol: string | null = null;
  if (category && categories) {
    const matchingCategory = categories[category];
    if (matchingCategory && matchingCategory.icon) {
      symbol = matchingCategory.icon;
    }
  }

  const node: Node = {
    id,
    category,
    title,
    depth,
    metadata,
    symbol,
  };
  return node;
};

const buildNodesAndEdges = (
  rawData: LeafNodeData | undefined,
  properties: FlowProperties | GraphProperties | HierarchyProperties = {},
  namedColors = {}
): NodesAndEdges => {
  if (!rawData || !rawData.columns || !rawData.rows) {
    return {
      nodes: [],
      root_nodes: {},
      edges: [],
      categories: {},
      next_color_index: 0,
    };
  }

  let categoryProperties = {};
  if (properties && properties.categories) {
    categoryProperties = properties.categories;
  }

  const id_col = getColumn(rawData.columns, "id");
  const from_col = getColumn(rawData.columns, "from_id");
  const to_col = getColumn(rawData.columns, "to_id");
  const title_col = getColumn(rawData.columns, "title");
  const category_col = getColumn(rawData.columns, "category");
  const depth_col = getColumn(rawData.columns, "depth");
  const metadata_col = getColumn(rawData.columns, "metadata");

  if (!id_col && !from_col && !to_col) {
    throw new Error("No node or edge rows defined in the dataset");
  }

  const node_lookup: NodeMap = {};
  const root_node_lookup: NodeMap = {};
  const edge_lookup: EdgeMap = {};
  const nodes: Node[] = [];
  const edges: Edge[] = [];
  const categories: CategoryMap = {};

  let contains_duplicate_edges = false;
  let colorIndex = 0;

  rawData.rows.forEach((row) => {
    const node_id: string | null = id_col ? row[id_col.name] : null;
    const from_id: string | null = from_col ? row[from_col.name] : null;
    const to_id: string | null = to_col ? row[to_col?.name] : null;
    const title: string | null = title_col ? row[title_col.name] : null;
    const category: string | null = category_col
      ? row[category_col.name]
      : null;
    const depth: number | null = depth_col ? row[depth_col.name] : null;
    const metadata: {} | null = metadata_col ? row[metadata_col.name] : null;

    if (category && !categories[category]) {
      const overrides = categoryProperties[category];
      const categorySettings = { color: null, depth: null, icon: null };
      if (overrides) {
        // @ts-ignore
        categorySettings.color =
          getColorOverride(overrides.color, namedColors) ||
          themeColors[colorIndex++];
        // @ts-ignore
        categorySettings.depth = has(overrides, "depth")
          ? overrides.depth
          : null;

        categorySettings.icon = has(overrides, "icon") ? overrides.icon : null;
      } else {
        // @ts-ignore
        categorySettings.color = themeColors[colorIndex++];
      }
      categories[category] = categorySettings;
    }

    // 4 types of row:
    //
    // id                  = node         1      1
    // from_id & id        = node & edge  1 2    3
    // id & to_id          = node & edge  1 4    5
    // from_id & to_id     = edge         2 4    6
    // id, from_id & to_id = node & edge  1 2 4  7

    const nodeAndEndMask =
      (node_id ? 1 : 0) + (from_id ? 2 : 0) + (to_id ? 4 : 0);
    const allowedNodeAndEdgeMasks = [1, 3, 5, 6, 7];

    // We must have at least a node id or an edge from_id / to_id pairing
    if (!allowedNodeAndEdgeMasks.includes(nodeAndEndMask)) {
      return new Error(
        `Encountered dataset row with no node or edge definition: ${JSON.stringify(
          row
        )}`
      );
    }

    // If this row is a node
    if (!!node_id) {
      const existingNode = node_lookup[node_id];

      if (!existingNode) {
        const node = createNode(
          node_id,
          title,
          category,
          depth,
          metadata,
          categories
        );
        node_lookup[node_id] = node;

        nodes.push(node);

        // Record this as a root node for now - we may remove that once we process the edges
        root_node_lookup[node_id] = node;
      } else {
        existingNode.title = title;
        existingNode.category = category;
        existingNode.depth = depth;
      }

      // If this has an edge from another node
      if (!!from_id && !to_id) {
        // If we've previously recorded this as a root node, remove it
        delete root_node_lookup[node_id];

        const existingNode = node_lookup[from_id];
        if (!existingNode) {
          const node = createNode(from_id);
          node_lookup[from_id] = node;

          nodes.push(node);

          // Record this as a root node for now - we may remove that once we process the edges
          root_node_lookup[from_id] = node;
        }

        const { edge, duplicate_edge } = recordEdge(
          edge_lookup,
          from_id,
          node_id
        );
        if (duplicate_edge) {
          contains_duplicate_edges = true;
        }
        edges.push(edge);
      }
      // Else if this has an edge to another node
      else if (!!to_id && !from_id) {
        // If we've previously recorded the target as a root node, remove it
        delete root_node_lookup[to_id];

        const existingNode = node_lookup[to_id];
        if (!existingNode) {
          const node = createNode(to_id);
          node_lookup[to_id] = node;

          nodes.push(node);
        }

        const { edge, duplicate_edge } = recordEdge(
          edge_lookup,
          node_id,
          to_id
        );
        if (duplicate_edge) {
          contains_duplicate_edges = true;
        }
        edges.push(edge);
      }
    }

    // If this row looks like an edge
    if (!!from_id && !!to_id) {
      // If we've previously recorded this as a root node, remove it
      delete root_node_lookup[to_id];

      // Record implicit nodes from edge definition
      const existingFromNode = node_lookup[from_id];
      if (!existingFromNode) {
        const node = createNode(from_id);
        node_lookup[from_id] = node;
        nodes.push(node);
        // Record this as a root node for now - we may remove that once we process the edges
        root_node_lookup[from_id] = node;
      }
      const existingToNode = node_lookup[to_id];
      if (!existingToNode) {
        const node = createNode(to_id);
        node_lookup[to_id] = node;
        nodes.push(node);
      }

      const { edge, duplicate_edge } = recordEdge(
        edge_lookup,
        from_id,
        to_id,
        title,
        category
      );
      if (duplicate_edge) {
        contains_duplicate_edges = true;
      }
      edges.push(edge);
    }
  });

  return {
    nodes,
    root_nodes: root_node_lookup,
    edges,
    categories,
    metadata: {
      has_multiple_roots: Object.keys(root_node_lookup).length > 1,
      contains_duplicate_edges,
    },
    next_color_index: colorIndex,
  };
};

const buildSankeyDataInputs = (nodesAndEdges: NodesAndEdges) => {
  const data: any[] = [];
  const links: any[] = [];
  const nodeDepths = {};

  nodesAndEdges.edges.forEach((edge) => {
    let categoryOverrides: Category = { color: null, icon: null };
    if (edge.category && nodesAndEdges.categories[edge.category]) {
      categoryOverrides = nodesAndEdges.categories[edge.category];
    }

    const existingFromDepth = nodeDepths[edge.from_id];
    if (!existingFromDepth) {
      nodeDepths[edge.from_id] = 0;
    }
    const existingToDepth = nodeDepths[edge.to_id];
    if (!existingToDepth) {
      nodeDepths[edge.to_id] = nodeDepths[edge.from_id] + 1;
    }
    links.push({
      source: edge.from_id,
      target: edge.to_id,
      value: 0.01,
      lineStyle: {
        color:
          categoryOverrides && categoryOverrides.color
            ? categoryOverrides.color
            : "target",
      },
    });
  });

  nodesAndEdges.nodes.forEach((node) => {
    let categoryOverrides;
    if (node.category && nodesAndEdges.categories[node.category]) {
      categoryOverrides = nodesAndEdges.categories[node.category];
    }
    const dataNode = {
      id: node.id,
      name: node.title,
      depth:
        node.depth !== null
          ? node.depth
          : has(categoryOverrides, "depth")
          ? categoryOverrides.depth
          : nodeDepths[node.id],
      itemStyle: {
        color:
          categoryOverrides && categoryOverrides.color
            ? categoryOverrides.color
            : themeColors[nodesAndEdges.next_color_index++],
      },
    };
    data.push(dataNode);
  });

  return {
    data,
    links,
  };

  // const objectData = rawData.rows.map((dataRow) => {
  //   const row: HierarchyDataRow = {
  //     parent: null,
  //     category: "",
  //     id: "",
  //     name: "",
  //   };
  //   for (let colIndex = 0; colIndex < rawData.columns.length; colIndex++) {
  //     const column = rawData.columns[colIndex];
  //     row[column.name] = dataRow[colIndex];
  //   }
  //
  //   if (row.category && !categories[row.category]) {
  //     let color;
  //     if (
  //       properties &&
  //       properties.categories &&
  //       properties.categories[row.category] &&
  //       properties.categories[row.category].color
  //     ) {
  //       color = properties.categories[row.category].color;
  //       colorIndex++;
  //     } else {
  //       color = themeColors[colorIndex++];
  //     }
  //     categories[row.category] = { color };
  //   }
  //
  //   if (!usedIds[row.id]) {
  //     builtData.push({
  //       ...row,
  //       itemStyle: {
  //         // @ts-ignore
  //         color: getColorOverride(categories[row.category].color, themeColors),
  //       },
  //     });
  //     usedIds[row.id] = true;
  //   }
  //   return row;
  // });
  // const edges: HierarchyDataRowEdge[] = [];
  // const edgeValues = {};
  // for (const d of objectData) {
  //   // TODO remove <null> after Kai fixes base64 issue and removes col string conversion
  //   if (d.parent === null || d.parent === "<null>") {
  //     d.parent = null;
  //     continue;
  //   }
  //   edges.push({ source: d.parent, target: d.id, value: 0.01 });
  //   edgeValues[d.parent] = (edgeValues[d.parent] || 0) + 0.01;
  // }
  // for (const e of edges) {
  //   var v = 0;
  //   if (edgeValues[e.target]) {
  //     for (const e2 of edges) {
  //       if (e.target === e2.source) {
  //         v += edgeValues[e2.target] || 0.01;
  //       }
  //     }
  //     e.value = v;
  //   }
  // }
  //
  // return {
  //   data: builtData,
  //   links: edges,
  // };
};

const buildGraphDataInputs = (nodesAndEdges: NodesAndEdges) => {
  const data: any[] = [];
  const links: any[] = [];
  const nodeDepths = {};

  nodesAndEdges.edges.forEach((edge) => {
    let categoryOverrides: Category = { color: null, icon: null };
    if (edge.category && nodesAndEdges.categories[edge.category]) {
      categoryOverrides = nodesAndEdges.categories[edge.category];
    }

    const existingFromDepth = nodeDepths[edge.from_id];
    if (!existingFromDepth) {
      nodeDepths[edge.from_id] = 0;
    }
    const existingToDepth = nodeDepths[edge.to_id];
    if (!existingToDepth) {
      nodeDepths[edge.to_id] = nodeDepths[edge.from_id] + 1;
    }
    links.push({
      source: edge.from_id,
      target: edge.to_id,
      value: 0.01,
      lineStyle: {
        color:
          categoryOverrides && categoryOverrides.color
            ? categoryOverrides.color
            : "target",
      },
    });
  });

  nodesAndEdges.nodes.forEach((node) => {
    let categoryOverrides;
    if (node.category && nodesAndEdges.categories[node.category]) {
      categoryOverrides = nodesAndEdges.categories[node.category];
    }
    const dataNode = {
      id: node.id,
      name: node.title,
      depth:
        node.depth !== null
          ? node.depth
          : has(categoryOverrides, "depth")
          ? categoryOverrides.depth
          : nodeDepths[node.id],
      itemStyle: {
        color:
          categoryOverrides && categoryOverrides.color
            ? categoryOverrides.color
            : themeColors[nodesAndEdges.next_color_index++],
      },
      metadata: node.metadata,
      symbol: node.symbol,
    };
    data.push(dataNode);
  });

  return {
    data,
    links,
  };

  // const objectData = rawData.rows.map((dataRow) => {
  //   const row: HierarchyDataRow = {
  //     parent: null,
  //     category: "",
  //     id: "",
  //     name: "",
  //   };
  //   for (let colIndex = 0; colIndex < rawData.columns.length; colIndex++) {
  //     const column = rawData.columns[colIndex];
  //     row[column.name] = dataRow[colIndex];
  //   }
  //
  //   if (row.category && !categories[row.category]) {
  //     let color;
  //     if (
  //       properties &&
  //       properties.categories &&
  //       properties.categories[row.category] &&
  //       properties.categories[row.category].color
  //     ) {
  //       color = properties.categories[row.category].color;
  //       colorIndex++;
  //     } else {
  //       color = themeColors[colorIndex++];
  //     }
  //     categories[row.category] = { color };
  //   }
  //
  //   if (!usedIds[row.id]) {
  //     builtData.push({
  //       ...row,
  //       itemStyle: {
  //         // @ts-ignore
  //         color: getColorOverride(categories[row.category].color, themeColors),
  //       },
  //     });
  //     usedIds[row.id] = true;
  //   }
  //   return row;
  // });
  // const edges: HierarchyDataRowEdge[] = [];
  // const edgeValues = {};
  // for (const d of objectData) {
  //   // TODO remove <null> after Kai fixes base64 issue and removes col string conversion
  //   if (d.parent === null || d.parent === "<null>") {
  //     d.parent = null;
  //     continue;
  //   }
  //   edges.push({ source: d.parent, target: d.id, value: 0.01 });
  //   edgeValues[d.parent] = (edgeValues[d.parent] || 0) + 0.01;
  // }
  // for (const e of edges) {
  //   var v = 0;
  //   if (edgeValues[e.target]) {
  //     for (const e2 of edges) {
  //       if (e.target === e2.source) {
  //         v += edgeValues[e2.target] || 0.01;
  //       }
  //     }
  //     e.value = v;
  //   }
  // }
  //
  // return {
  //   data: builtData,
  //   links: edges,
  // };
};

interface Item {
  [key: string]: any;
}

interface TreeItem {
  [key: string]: Item | TreeItem[] | any;
}

// Taken from https://github.com/philipstanislaus/performant-array-to-tree
const nodesAndEdgesToTree = (nodesAndEdges: NodesAndEdges): TreeItem[] => {
  // const rootParentIds = { "": true };

  // the resulting unflattened tree
  // const rootItems: TreeItem[] = [];

  // stores all already processed items with their ids as key so we can easily look them up
  const lookup: { [id: string]: TreeItem } = {};

  // stores all item ids that have not been added to the resulting unflattened tree yet
  // this is an opt-in property, since it has a slight runtime overhead
  // const orphanIds: null | Set<string | number> = new Set();

  let colorIndex = 0;

  // Add in the nodes to the lookup
  for (const node of nodesAndEdges.nodes) {
    // look whether item already exists in the lookup table
    if (!lookup[node.id]) {
      // item is not yet there, so add a preliminary item (its data will be added later)
      lookup[node.id] = { children: [] };
    }

    let color;
    if (node.category && nodesAndEdges.categories[node.category]) {
      const categoryOverrides = nodesAndEdges.categories[node.category];
      if (categoryOverrides.color) {
        color = categoryOverrides.color;
        colorIndex++;
      } else {
        color = themeColors[colorIndex++];
      }
    }

    lookup[node.id] = {
      ...node,
      name: node.title,
      itemStyle: {
        color,
      },
      children: lookup[node.id].children,
    };
  }
  // Fill in the children with the edge relationships
  for (const edge of nodesAndEdges.edges) {
    const childId = edge.to_id;
    const parentId = edge.from_id;

    // look whether the parent already exists in the lookup table
    if (!lookup[parentId]) {
      // parent is not yet there, so add a preliminary parent (its data will be added later)
      lookup[parentId] = { children: [] };
    }

    const childItem = lookup[childId];

    // add the current item to the parent
    lookup[parentId].children.push(childItem);
  }
  return Object.values(lookup).filter(
    (node) => nodesAndEdges.root_nodes[node.id]
  );

  // idea of this loop:
  // whenever an item has a parent, but the parent is not yet in the lookup object, we store a preliminary parent
  // in the lookup object and fill it with the data of the parent later
  // if an item has no parentId, add it as a root element to rootItems
  // for (const rawRow of rawData.rows) {
  //   const row: HierarchyDataRow = {
  //     parent: null,
  //     category: "",
  //     id: "",
  //     name: "",
  //   };
  //   for (let colIndex = 0; colIndex < rawData.columns.length; colIndex++) {
  //     const column = rawData.columns[colIndex];
  //     row[column.name] = rawRow[colIndex];
  //   }
  //
  //   const itemId = row.id;
  //   const parentId = row.parent;
  //
  //   if (rootParentIds[itemId]) {
  //     throw new Error(
  //       `The row contains a node whose parent both exists in another node and is in ` +
  //         `\`rootParentIds\` (\`itemId\`: "${itemId}", \`rootParentIds\`: ${Object.keys(
  //           rootParentIds
  //         )
  //           .map((r) => `"${r}"`)
  //           .join(", ")}).`
  //     );
  //   }
  //
  //   // look whether item already exists in the lookup table
  //   if (!Object.prototype.hasOwnProperty.call(lookup, itemId)) {
  //     // item is not yet there, so add a preliminary item (its data will be added later)
  //     lookup[itemId] = { children: [] };
  //   }
  //
  //   // if we track orphans, delete this item from the orphan set if it is in it
  //   if (orphanIds) {
  //     orphanIds.delete(itemId);
  //   }
  //
  //   // add the current item's data to the item in the lookup table
  //
  //   lookup[itemId] = {
  //     ...row,
  //     children: lookup[itemId].children,
  //   };
  //
  //   const treeItem = lookup[itemId];
  //
  //   if (
  //     parentId === null ||
  //     parentId === undefined ||
  //     rootParentIds[parentId]
  //   ) {
  //     // is a root item
  //     rootItems.push(treeItem);
  //   } else {
  //     // has a parent
  //
  //     // look whether the parent already exists in the lookup table
  //     if (!Object.prototype.hasOwnProperty.call(lookup, parentId)) {
  //       // parent is not yet there, so add a preliminary parent (its data will be added later)
  //       lookup[parentId] = { children: [] };
  //
  //       // if we track orphans, add the generated parent to the orphan list
  //       if (orphanIds) {
  //         orphanIds.add(parentId);
  //       }
  //     }
  //
  //     // add the current item to the parent
  //     lookup[parentId].children.push(treeItem);
  //   }
  // }

  // if (orphanIds?.size) {
  //   throw new Error(
  //     `The items array contains orphans that point to the following parentIds: ` +
  //       `[${Array.from(
  //         orphanIds
  //       )}]. These parentIds do not exist in the items array. Hint: prevent orphans to result ` +
  //       `in an error by passing the following option: { throwIfOrphans: false }`
  //   );
  // }

  // return rootItems;
};

const buildTreeDataInputs = (nodesAndEdges: NodesAndEdges) => {
  const tree = nodesAndEdgesToTree(nodesAndEdges);
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

const getColorOverride = (colorOverride, namedThemeColors) => {
  if (colorOverride === "alert") {
    return namedThemeColors.alert;
  }
  if (colorOverride === "info") {
    return namedThemeColors.info;
  }
  if (colorOverride === "ok") {
    return namedThemeColors.ok;
  }
  return colorOverride;
};

export {
  adjustMinValue,
  adjustMaxValue,
  buildChartDataset,
  buildGraphDataInputs,
  buildNodesAndEdges,
  buildSankeyDataInputs,
  buildTreeDataInputs,
  getColorOverride,
  isNumericCol,
  themeColors,
  toEChartsType,
};
