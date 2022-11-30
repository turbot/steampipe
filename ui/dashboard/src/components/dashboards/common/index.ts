import has from "lodash/has";
import isEmpty from "lodash/isEmpty";
import {
  Category,
  CategoryMap,
  Edge,
  EdgeMap,
  KeyValuePairs,
  Node,
  NodeCategoryMap,
  NodeMap,
  NodesAndEdges,
} from "./types";
import { ChartProperties, ChartTransform, ChartType } from "../charts/types";
import { DashboardRunState } from "../../../types";
import { ExpandedNodes } from "../graphs/common/useGraph";
import { FlowProperties, FlowType } from "../flows/types";
import { getColumn } from "../../../utils/data";
import { Graph, json } from "graphlib";
import { GraphProperties, GraphType, NodeAndEdgeData } from "../graphs/types";
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

const recordEdge = (
  edge_lookup,
  from_id: string,
  to_id: string,
  title: string | null = null,
  category: string | null = null,
  row_data: LeafNodeDataRow | null = null
) => {
  let duplicate_edge = false;
  // Find any existing edge
  const edge_id = `${from_id}_${to_id}`;
  const existingNode = edge_lookup[edge_id];

  const edge: Edge = {
    id: edge_id,
    from_id,
    to_id,
    title,
    category,
    row_data,
    isFolded: false,
  };

  if (existingNode) {
    duplicate_edge = true;
  } else {
    edge_lookup[edge_id] = edge;
  }

  return {
    edge,
    duplicate_edge,
  };
};

const createNode = (
  node_lookup,
  nodes_by_category,
  id: string,
  title: string | null = null,
  category: string | null = null,
  depth: number | null = null,
  row_data: LeafNodeDataRow | null = null,
  categories: CategoryMap = {},
  isFolded: boolean = false
) => {
  let symbol: string | null = null;
  let href: string | null = null;
  if (category && categories) {
    const matchingCategory = categories[category];
    if (matchingCategory && matchingCategory.icon) {
      symbol = matchingCategory.icon;
    }
    if (matchingCategory && matchingCategory.href) {
      href = matchingCategory.href;
    }
  }

  const node: Node = {
    id,
    category,
    title,
    depth,
    row_data,
    symbol,
    href,
    isFolded,
  };
  node_lookup[id] = node;

  if (category) {
    nodes_by_category[category] = nodes_by_category[category] || {};
    nodes_by_category[category][id] = node;
  }
  return node;
};

const getCategoriesWithFold = (categories: CategoryMap): CategoryMap => {
  if (!categories) {
    return {};
  }
  return Object.entries(categories)
    .filter(([_, info]) => !!info.fold)
    .reduce((res, [category, info]) => {
      res[category] = info;
      return res;
    }, {});
};

const foldNodesAndEdges = (
  nodesAndEdges: NodesAndEdges,
  expandedNodes: ExpandedNodes = {}
): NodesAndEdges => {
  const categoriesWithFold = getCategoriesWithFold(nodesAndEdges.categories);

  if (isEmpty(categoriesWithFold)) {
    return nodesAndEdges;
  }

  const newNodesAndEdges = {
    ...nodesAndEdges,
  };

  const graph = json.read(json.write(nodesAndEdges.graph));

  for (const [category, info] of Object.entries(categoriesWithFold)) {
    // Keep track of the number of folded nodes we've added
    let foldedNodeCount = 0;

    // Find all nodes of this given category
    const nodesForCategory = nodesAndEdges.nodeCategoryMap[category];

    // If we have no nodes for this category, continue
    if (!nodesForCategory) {
      continue;
    }

    // If the number of nodes for this category is less than the threshold, it's
    // not possible that any would require folding, regardless of the graph structure
    const categoryNodesById = Object.entries(nodesForCategory);

    if (categoryNodesById.length < (info.fold?.threshold || 0)) {
      continue;
    }

    // Now we're here we know that we have enough nodes of this category in the
    // graph that it "might" be possible to fold, but we'll examine the
    // node and edge structure now to determine that

    const categoryEdgeGroupings: KeyValuePairs = {};

    // Iterate over the category nodes
    for (const [, node] of categoryNodesById) {
      let sourceNodes: string[] = [];
      let targetNodes: string[] = [];

      // Get all the in edges to this node
      const inEdges = graph.inEdges(node.id);

      // Get all the out edges from this node
      const outEdges = graph.outEdges(node.id);

      // Record the nodes pointing to this node
      for (const inEdge of inEdges || []) {
        sourceNodes.push(inEdge.v);
      }

      // Record the nodes this node points to
      for (const outEdge of outEdges || []) {
        targetNodes.push(outEdge.w);
      }

      // Sort to ensure consistent
      sourceNodes = sourceNodes.sort();
      targetNodes = targetNodes.sort();

      // Build a key that we can uniquely identify each unique combo category / source nodes / target nodes
      // and record all the nodes for that key. If we have any keys that have >= fold threshold, fold them
      const categoryGroupingKey = `category:${node.category}`;
      const edgeSourceGroupingKey =
        sourceNodes.length > 0 ? `source:${sourceNodes.join(",")}` : null;
      const edgeTargetGroupingKey =
        targetNodes.length > 0 ? `target:${targetNodes.join(",")}` : null;
      const edgeGroupingKey = `${categoryGroupingKey}${
        edgeSourceGroupingKey ? `_${edgeSourceGroupingKey}` : ""
      }${edgeTargetGroupingKey ? `_${edgeTargetGroupingKey}` : ""}`;
      categoryEdgeGroupings[edgeGroupingKey] = categoryEdgeGroupings[
        edgeGroupingKey
      ] || {
        category: info,
        threshold: info.fold?.threshold,
        nodes: [],
        source: sourceNodes,
        target: targetNodes,
      };
      categoryEdgeGroupings[edgeGroupingKey].nodes.push(node);
    }

    // Find any nodes that can be folded
    for (const [, groupingInfo] of Object.entries(categoryEdgeGroupings)
      // @ts-ignore
      .filter(
        ([_, g]) =>
          g.threshold !== null &&
          g.threshold !== undefined &&
          g.nodes.length >= g.threshold
      )) {
      const removedNodes: any[] = [];

      // Create a structure to capture the category and title of each edge that
      // is being folded into this node. Later, if they are all the same, we can
      // use that same category and title for the new folded edge.
      const deletedSourceEdges = { categories: {}, titles: {} };
      const deletedTargetEdges = { categories: {}, titles: {} };

      // We want to fold nodes that are not expanded
      for (const node of groupingInfo.nodes) {
        // This node is expanded, don't fold it
        if (expandedNodes[node.id]) {
          continue;
        }

        // Remove this node
        graph.removeNode(node.id);
        delete newNodesAndEdges.nodeMap[node.id];
        delete newNodesAndEdges.nodeCategoryMap[category][node.id];
        // Remove edges pointing to this node
        for (const sourceNode of groupingInfo.source) {
          const sourceEdgeKey = `${sourceNode}_${node.id}`;
          const sourceEdge = newNodesAndEdges.edgeMap[sourceEdgeKey];
          const sourceEdgeTitle = sourceEdge.title || "none";
          const sourceEdgeCategory = sourceEdge.category || "none";
          deletedSourceEdges.categories[sourceEdgeCategory] =
            deletedSourceEdges.categories[sourceEdgeCategory] || 0;
          deletedSourceEdges.categories[sourceEdgeCategory]++;
          deletedSourceEdges.titles[sourceEdgeTitle] =
            deletedSourceEdges.titles[sourceEdgeTitle] || 0;
          deletedSourceEdges.titles[sourceEdgeTitle]++;
          delete newNodesAndEdges.edgeMap[sourceEdgeKey];
          graph.removeEdge(sourceNode, node.id);
        }
        // Remove edges coming from this node
        for (const targetNode of groupingInfo.target) {
          const targetEdgeKey = `${node.id}_${targetNode}`;
          const targetEdge = newNodesAndEdges.edgeMap[targetEdgeKey];
          const targetEdgeTitle =
            targetEdge.title || targetEdge.category || "none";
          const targetEdgeCategory = targetEdge.category || "none";
          deletedTargetEdges.categories[targetEdgeCategory] =
            deletedTargetEdges.categories[targetEdgeCategory] || 0;
          deletedTargetEdges.categories[targetEdgeCategory]++;
          deletedTargetEdges.titles[targetEdgeTitle] =
            deletedTargetEdges.titles[targetEdgeTitle] || 0;
          deletedTargetEdges.titles[targetEdgeTitle]++;
          delete newNodesAndEdges.edgeMap[targetEdgeKey];
          graph.removeEdge(node.id, targetNode);
        }
        removedNodes.push({ id: node.id, title: node.title });
      }

      // Now let's add a folded node
      if (removedNodes.length > 0) {
        const foldedNode = {
          id: `fold-${category}-${++foldedNodeCount}`,
          category,
          icon: info.fold?.icon,
          title: info.fold?.title ? info.fold.title : null,
          isFolded: true,
          foldedNodes: removedNodes,
          row_data: null,
          href: null,
          depth: null,
          symbol: null,
        };
        graph.setNode(foldedNode.id);
        newNodesAndEdges.nodeCategoryMap[category][foldedNode.id] = foldedNode;
        newNodesAndEdges.nodeMap[foldedNode.id] = foldedNode;

        // We want to add the color and category if all edges to this node have a common color or category
        const deletedSourceEdgeCategoryKeys = Object.keys(
          deletedSourceEdges.categories
        );
        const deletedSourceEdgeTitleKeys = Object.keys(
          deletedSourceEdges.titles
        );
        const sourceEdgeCategory =
          deletedSourceEdgeCategoryKeys.length === 1 &&
          deletedSourceEdgeCategoryKeys[0] !== "none"
            ? deletedSourceEdgeCategoryKeys[0]
            : null;
        const sourceEdgeTitle =
          deletedSourceEdgeTitleKeys.length === 1 &&
          deletedSourceEdgeTitleKeys[0] !== "none"
            ? deletedSourceEdgeTitleKeys[0]
            : null;

        // Add the source edges back to the folded node
        for (const sourceNode of groupingInfo.source) {
          graph.setEdge(sourceNode, foldedNode.id);
          const edge: Edge = {
            id: `${sourceNode}_${foldedNode.id}`,
            from_id: sourceNode,
            to_id: foldedNode.id,
            category: sourceEdgeCategory,
            title: sourceEdgeTitle,
            isFolded: true,
            row_data: null,
          };
          newNodesAndEdges.edgeMap[edge.id] = edge;
        }

        // We want to add the category and title if all edges from this node have a common category or title
        const deletedTargetEdgeCategoryKeys = Object.keys(
          deletedTargetEdges.categories
        );
        const deletedTargetEdgeTitleKeys = Object.keys(
          deletedTargetEdges.titles
        );
        const targetEdgeCategory =
          deletedTargetEdgeCategoryKeys.length === 1 &&
          deletedTargetEdgeCategoryKeys[0] !== "none"
            ? deletedTargetEdgeCategoryKeys[0]
            : null;
        const targetEdgeTitle =
          deletedTargetEdgeTitleKeys.length === 1 &&
          deletedTargetEdgeTitleKeys[0] !== "none"
            ? deletedTargetEdgeTitleKeys[0]
            : null;

        // Add the target edges back from the folded node
        for (const targetNode of groupingInfo.target) {
          graph.setEdge(foldedNode.id, targetNode);
          const edge = {
            id: `${foldedNode.id}_${targetNode}`,
            from_id: foldedNode.id,
            to_id: targetNode,
            category: targetEdgeCategory,
            title: targetEdgeTitle,
          };
          newNodesAndEdges.edgeMap[edge.id] = edge;
        }
      }
    }
  }

  return {
    ...newNodesAndEdges,
    nodes: graph.nodes().map((nodeId) => newNodesAndEdges.nodeMap[nodeId]),
    edges: graph
      .edges()
      .map((edgeObj) => newNodesAndEdges.edgeMap[`${edgeObj.v}_${edgeObj.w}`]),
  };
};

const buildNodesAndEdges = (
  categories: CategoryMap = {},
  rawData: NodeAndEdgeData | undefined,
  properties: FlowProperties | GraphProperties | HierarchyProperties = {},
  namedThemeColors = {},
  defaultCategoryColor = true
): NodesAndEdges => {
  if (!rawData || !rawData.columns || !rawData.rows) {
    return {
      graph: new Graph(),
      nodes: [],
      edges: [],
      nodeCategoryMap: {},
      nodeMap: {},
      edgeMap: {},
      root_nodes: {},
      categories: {},
      next_color_index: 0,
    };
  }

  const graph = new Graph({ directed: true });

  let categoryProperties = {};
  if (properties && properties.categories) {
    categoryProperties = properties.categories;
  }

  const id_col = getColumn(rawData.columns, "id");
  const from_col = getColumn(rawData.columns, "from_id");
  const to_col = getColumn(rawData.columns, "to_id");

  if (!id_col && !from_col && !to_col) {
    return {
      graph: new Graph(),
      nodes: [],
      edges: [],
      nodeCategoryMap: {},
      nodeMap: {},
      edgeMap: {},
      root_nodes: {},
      categories: {},
      next_color_index: 0,
    };
  }

  const node_lookup: NodeMap = {};
  const root_node_lookup: NodeMap = {};
  const nodes_by_category: NodeCategoryMap = {};
  const edge_lookup: EdgeMap = {};
  const nodes: Node[] = [];
  const edges: Edge[] = [];

  let contains_duplicate_edges = false;
  let colorIndex = 0;

  rawData.rows.forEach((row) => {
    const node_id: string | null =
      has(row, "id") && row.id !== null && row.id !== undefined
        ? row.id.toString()
        : null;
    const from_id: string | null =
      has(row, "from_id") && row.from_id !== null && row.from_id !== undefined
        ? row.from_id.toString()
        : null;
    const to_id: string | null =
      has(row, "to_id") && row.to_id !== null && row.to_id !== undefined
        ? row.to_id.toString()
        : null;
    const title: string | null = row.title || null;
    const category: string | null = row.category || null;
    const depth: number | null =
      typeof row.depth === "number" ? row.depth : null;

    if (category && !categories[category]) {
      const overrides = categoryProperties[category];
      const categorySettings: Category = {};
      if (overrides) {
        const overrideColor = getColorOverride(
          overrides.color,
          namedThemeColors
        );
        // @ts-ignore
        categorySettings.color = overrideColor
          ? overrideColor
          : defaultCategoryColor
          ? themeColors[colorIndex++]
          : null;
        if (has(overrides, "depth")) {
          categorySettings.depth = overrides.depth;
        }
        if (has(overrides, "fields")) {
          categorySettings.fields = overrides.fields;
        }
        if (has(overrides, "icon")) {
          categorySettings.icon = overrides.icon;
        }
        if (has(overrides, "href")) {
          categorySettings.href = overrides.href;
        }
        if (has(overrides, "fold")) {
          categorySettings.fold = overrides.fold;
        }
      } else {
        // @ts-ignore
        categorySettings.color = defaultCategoryColor
          ? themeColors[colorIndex++]
          : null;
      }
      categories[category] = categorySettings;
    }

    // 5 types of row:
    //
    // id                  = node         1      1
    // from_id & id        = node & edge  1 2    3
    // id & to_id          = node & edge  1 4    5
    // from_id & to_id     = edge         2 4    6
    // id, from_id & to_id = node & edge  1 2 4  7

    const nodeAndEdgeMask =
      (node_id ? 1 : 0) + (from_id ? 2 : 0) + (to_id ? 4 : 0);
    const allowedNodeAndEdgeMasks = [1, 3, 5, 6, 7];

    // We must have at least a node id or an edge from_id / to_id pairing
    if (!allowedNodeAndEdgeMasks.includes(nodeAndEdgeMask)) {
      return new Error(
        `Encountered dataset row with no node or edge definition: ${JSON.stringify(
          row
        )}`
      );
    }

    // If this row is a node
    if (!!node_id) {
      const existingNode = node_lookup[node_id];

      // Even if the node already existed, it will only have minimal info, as it
      // could only have been created implicitly through an edge definition, so
      // build a full node and update it
      const node = createNode(
        node_lookup,
        nodes_by_category,
        node_id,
        title,
        category,
        depth,
        row,
        categories
      );

      // Ensure that any existing references to this node are also updated
      if (existingNode) {
        const nodeIndex = nodes.findIndex((n) => n.id === node.id);
        if (nodeIndex >= 0) {
          nodes[nodeIndex] = node;
        }
        if (root_node_lookup[node.id]) {
          root_node_lookup[node.id] = node;
        }
      } else {
        graph.setNode(node_id);
        nodes.push(node);

        // Record this as a root node for now - we may remove that once we process the edges
        root_node_lookup[node_id] = node;
      }

      // If this has an edge from another node
      if (!!from_id && !to_id) {
        // If we've previously recorded this as a root node, remove it
        delete root_node_lookup[node_id];

        const existingNode = node_lookup[from_id];
        if (!existingNode) {
          const node = createNode(
            node_lookup,
            nodes_by_category,
            from_id,
            null,
            null,
            null,
            null,
            {}
          );
          graph.setNode(from_id);
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
        graph.setEdge(from_id, node_id);
        edges.push(edge);
      }
      // Else if this has an edge to another node
      else if (!!to_id && !from_id) {
        // If we've previously recorded the target as a root node, remove it
        delete root_node_lookup[to_id];

        const existingNode = node_lookup[to_id];
        if (!existingNode) {
          const node = createNode(
            node_lookup,
            nodes_by_category,
            to_id,
            null,
            null,
            null,
            null,
            {}
          );
          graph.setNode(to_id);
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
        graph.setEdge(node_id, to_id);
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
        const node = createNode(
          node_lookup,
          nodes_by_category,
          from_id,
          null,
          null,
          null,
          null,
          {}
        );
        graph.setNode(from_id);
        nodes.push(node);
        // Record this as a root node for now - we may remove that once we process the edges
        root_node_lookup[from_id] = node;
      }
      const existingToNode = node_lookup[to_id];
      if (!existingToNode) {
        const node = createNode(
          node_lookup,
          nodes_by_category,
          to_id,
          null,
          null,
          null,
          null,
          {}
        );
        graph.setNode(to_id);
        nodes.push(node);
      }

      const { edge, duplicate_edge } = recordEdge(
        edge_lookup,
        from_id,
        to_id,
        title,
        category,
        nodeAndEdgeMask === 6 ? row : null
      );
      if (duplicate_edge) {
        contains_duplicate_edges = true;
      }
      graph.setEdge(from_id, to_id);
      edges.push(edge);
    }
  });

  return {
    graph,
    nodes,
    edges,
    nodeCategoryMap: nodes_by_category,
    nodeMap: node_lookup,
    edgeMap: edge_lookup,
    root_nodes: root_node_lookup,
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
    let categoryOverrides: Category = {};
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
            : themeColors[
                has(nodesAndEdges, "next_color_index")
                  ? // @ts-ignore
                    nodesAndEdges.next_color_index++
                  : 0
              ],
      },
    };
    data.push(dataNode);
  });

  return {
    data,
    links,
  };
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
  buildNodesAndEdges,
  buildSankeyDataInputs,
  buildTreeDataInputs,
  foldNodesAndEdges,
  getColorOverride,
  isNumericCol,
  themeColors,
  toEChartsType,
};
