import has from "lodash/has";
import isNumber from "lodash/isNumber";
import useChartThemeColors from "../../../hooks/useChartThemeColors";
import {
  Category,
  CategoryMap,
  EdgeProperties,
  KeyValueStringPairs,
  NodeAndEdgeProperties,
  NodeProperties,
} from "./types";
import {
  DashboardRunState,
  DependencyPanelProperties,
  PanelsMap,
} from "../../../types";
import { getColorOverride } from "./index";
import {
  NodeAndEdgeData,
  NodeAndEdgeDataColumn,
  NodeAndEdgeDataFormat,
  NodeAndEdgeDataRow,
  NodeAndEdgeStatus,
  WithStatusMap,
} from "../graphs/types";
import { useDashboard } from "../../../hooks/useDashboard";
import { useMemo } from "react";

// Categories may be sourced from a node, an edge, a flow, a graph or a hierarchy
// A node or edge can define exactly 1 category, which covers all rows of data that don't define a category
// Any node/edge/flow/graph/hierarchy data row can define a category - in that event, the category must come from the category map for the composing resource

const getNodeAndEdgeDataFormat = (
  properties: NodeAndEdgeProperties | undefined
): NodeAndEdgeDataFormat => {
  if (!properties) {
    return "LEGACY";
  }

  if (!properties.nodes && !properties.edges) {
    return "LEGACY";
  }

  if (
    (properties.nodes && properties.nodes.length > 0) ||
    (properties.edges && properties.edges.length > 0)
  ) {
    return "NODE_AND_EDGE";
  }

  return "LEGACY";
};

const addColumnsForResource = (
  columns: NodeAndEdgeDataColumn[],
  data: NodeAndEdgeData
): NodeAndEdgeDataColumn[] => {
  // Get a union of all the columns across all nodes
  const newColumns = [...columns];
  for (const column of data.columns || []) {
    if (newColumns.some((c) => c.name === column.name)) {
      continue;
    }
    newColumns.push(column);
  }
  return newColumns;
};

const populateCategoryWithDefaults = (
  category: Category,
  themeColors: KeyValueStringPairs
): Category => {
  return {
    name: category.name,
    color: getColorOverride(category.color, themeColors),
    depth: category.depth,
    properties: category.properties,
    fold: {
      threshold:
        category.fold && isNumber(category.fold.threshold)
          ? category.fold.threshold
          : 3,
      title: category.fold?.title || category.title,
      icon: category.fold?.icon || category.icon,
    },
    href: category.href,
    icon: category.icon,
    title: category.title,
  };
};

const emptyPanels: PanelsMap = {};

const addPanelWithsStatus = (
  panelsMap: PanelsMap,
  dependencies: string[] | undefined,
  withLookup: KeyValueStringPairs,
  withStatuses: WithStatusMap
) => {
  for (const dependency of dependencies || []) {
    // If we've already logged the status of this with, carry on
    if (withLookup[dependency] || dependency.indexOf(".with.") === -1) {
      continue;
    }

    const dependencyPanel = panelsMap[dependency];
    if (!dependencyPanel) {
      continue;
    }
    const dependencyPanelProperties =
      dependencyPanel.properties as DependencyPanelProperties;
    withLookup[dependency] = dependencyPanelProperties.name;
    withStatuses[dependencyPanelProperties.name] = {
      id: dependencyPanelProperties.name,
      title: dependencyPanel.title,
      state: dependencyPanel.status || "initialized",
      error: dependencyPanel.error,
    };
  }
};

// This function will normalise both the legacy and node/edge data formats into a data table.
// In the node/edge approach, the data will be spread out across the node and edge resources
// until the flow/graph/hierarchy has completed, at which point we'll have a populated data
// table in the parent resource.
const useNodeAndEdgeData = (
  data: NodeAndEdgeData | undefined,
  properties: NodeAndEdgeProperties | undefined,
  status: DashboardRunState
) => {
  const { panelsMap } = useDashboard();
  const themeColors = useChartThemeColors();
  const dataFormat = getNodeAndEdgeDataFormat(properties);
  const panels = useMemo(() => {
    if (dataFormat === "LEGACY") {
      return emptyPanels;
    }
    return panelsMap;
  }, [panelsMap, dataFormat]);

  return useMemo(() => {
    if (dataFormat === "LEGACY") {
      if (status === "complete") {
        const categories: CategoryMap = {};

        // Set defaults on categories
        for (const [name, category] of Object.entries(
          properties?.categories || {}
        )) {
          categories[name] = populateCategoryWithDefaults(
            category,
            themeColors
          );
        }

        return data ? { categories, data, dataFormat, properties } : null;
      }
      return null;
    }

    // We've now established that it's a NODE_AND_EDGE format data set, so let's build
    // what we need from the component parts

    let columns: NodeAndEdgeDataColumn[] = [];
    let rows: NodeAndEdgeDataRow[] = [];
    const categories: CategoryMap = {};
    const withNameLookup: KeyValueStringPairs = {};

    // Add flow/graph/hierarchy level categories
    for (const [name, category] of Object.entries(
      properties?.categories || {}
    )) {
      categories[name] = populateCategoryWithDefaults(category, themeColors);
    }

    const missingNodes = {};
    const missingEdges = {};
    const nodeAndEdgeStatus: NodeAndEdgeStatus = {
      withs: {},
      nodes: [],
      edges: [],
    };
    const nodeIdLookup = {};

    // Loop over all the node names and check out their respective panel in the panels map
    for (const nodePanelName of properties?.nodes || []) {
      const panel = panels[nodePanelName];

      // Capture missing panels - we'll deal with that after
      if (!panel) {
        missingNodes[nodePanelName] = true;
        continue;
      }

      // Capture the status of any with blocks that this node depends on
      addPanelWithsStatus(
        panels,
        panel.dependencies,
        withNameLookup,
        nodeAndEdgeStatus.withs
      );

      const typedPanelData = (panel.data || {}) as NodeAndEdgeData;
      columns = addColumnsForResource(columns, typedPanelData);
      const nodeProperties = (panel.properties || {}) as NodeProperties;
      const nodeDataRows = typedPanelData.rows || [];

      // Capture the status of this node resource
      nodeAndEdgeStatus.nodes.push({
        id: panel.title || nodeProperties.name,
        state: panel.status || "initialized",
        category: nodeProperties.category,
        error: panel.error,
        title: panel.title,
        dependencies: panel.dependencies,
      });

      let nodeCategory: Category | null = null;
      let nodeCategoryId: string = "";
      if (nodeProperties.category) {
        nodeCategory = populateCategoryWithDefaults(
          nodeProperties.category,
          themeColors
        );
        nodeCategoryId = `node.${nodePanelName}.${nodeCategory.name}`;
        categories[nodeCategoryId] = nodeCategory;
      }

      // Loop over each row and ensure we have the correct category information set for it
      for (const row of nodeDataRows) {
        // Ensure each row has an id
        if (row.id === null || row.id === undefined) {
          continue;
        }

        const updatedRow = { ...row };

        // Ensure the row has a title and populate from the node if not set
        if (!updatedRow.title && panel.title) {
          updatedRow.title = panel.title;
        }

        // Capture the ID of each row
        nodeIdLookup[row.id.toString()] = row;

        // If the row specifies a category and it's the same now as the node specified,
        // then update the category to the artificial node category ID
        if (updatedRow.category && nodeCategory?.name === updatedRow.category) {
          updatedRow.category = nodeCategoryId;
        }
        // Else if the row has a category, but we don't know about it, clear it
        else if (updatedRow.category && !categories[updatedRow.category]) {
          updatedRow.category = undefined;
        } else if (!updatedRow.category && nodeCategoryId) {
          updatedRow.category = nodeCategoryId;
        }
        rows.push(updatedRow);
      }
    }

    // Loop over all the edge names and check out their respective panel in the panels map
    for (const edgePanelName of properties?.edges || []) {
      const panel = panels[edgePanelName];

      // Capture missing panels - we'll deal with that after
      if (!panel) {
        missingEdges[edgePanelName] = true;
        continue;
      }

      // Capture the status of any with blocks that this edge depends on
      addPanelWithsStatus(
        panels,
        panel.dependencies,
        withNameLookup,
        nodeAndEdgeStatus.withs
      );

      const typedPanelData = (panel.data || {}) as NodeAndEdgeData;
      columns = addColumnsForResource(columns, typedPanelData);
      const edgeProperties = (panel.properties || {}) as EdgeProperties;

      // Capture the status of this edge resource
      nodeAndEdgeStatus.edges.push({
        id: panel.title || edgeProperties.name,
        state: panel.status || "initialized",
        category: edgeProperties.category,
        error: panel.error,
        title: panel.title,
        dependencies: panel.dependencies,
      });

      let edgeCategory: Category | null = null;
      let edgeCategoryId: string = "";
      if (edgeProperties.category) {
        edgeCategory = populateCategoryWithDefaults(
          edgeProperties.category,
          themeColors
        );
        edgeCategoryId = `edge.${edgePanelName}.${edgeCategory.name}`;
        categories[edgeCategoryId] = edgeCategory;
      }

      for (const row of typedPanelData.rows || []) {
        // Ensure the node this edge points to exists in the data set
        // @ts-ignore
        const from_id =
          has(row, "from_id") &&
          row.from_id !== null &&
          row.from_id !== undefined
            ? row.from_id.toString()
            : null;
        // @ts-ignore
        const to_id =
          has(row, "to_id") && row.to_id !== null && row.to_id !== undefined
            ? row.to_id.toString()
            : null;
        if (
          !from_id ||
          !to_id ||
          !nodeIdLookup[from_id] ||
          !nodeIdLookup[to_id]
        ) {
          continue;
        }

        const updatedRow = { ...row };

        // Ensure the row has a title and populate from the edge if not set
        if (!updatedRow.title && panel.title) {
          updatedRow.title = panel.title;
        }

        // If the row specifies a category and it's the same now as the edge specified,
        // then update the category to the artificial edge category ID
        if (updatedRow.category && edgeCategory?.name === updatedRow.category) {
          updatedRow.category = edgeCategoryId;
        }
        // Else if the row has a category, but we don't know about it, clear it
        else if (updatedRow.category && !categories[updatedRow.category]) {
          updatedRow.category = undefined;
        } else if (!updatedRow.category && edgeCategoryId) {
          updatedRow.category = edgeCategoryId;
        }
        rows.push(updatedRow);
      }
    }

    return {
      categories,
      data: { columns, rows },
      dataFormat,
      properties,
      status: nodeAndEdgeStatus,
    };
  }, [data, dataFormat, panels, properties, status, themeColors]);
};

export default useNodeAndEdgeData;

export { getNodeAndEdgeDataFormat };
