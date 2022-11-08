import has from "lodash/has";
import set from "lodash/set";
import { DashboardRunState } from "../../../types";
import { EdgeProperties, NodeAndEdgeProperties, NodeProperties } from "./types";
import {
  NodeAndEdgeData,
  NodeAndEdgeDataColumn,
  NodeAndEdgeDataFormat,
  NodeAndEdgeDataRow,
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

const useNodeAndEdgeData = (
  data: NodeAndEdgeData | undefined,
  properties: NodeAndEdgeProperties | undefined,
  status: DashboardRunState
) => {
  const { panelsMap } = useDashboard();
  return useMemo(() => {
    if (getNodeAndEdgeDataFormat(properties) === "LEGACY") {
      if (status === "complete") {
        return data ? { data, properties } : null;
      }
      return null;
    }

    let newProperties = properties;
    const nodeIdLookup = {};
    const columns: NodeAndEdgeDataColumn[] = [];
    const rows: NodeAndEdgeDataRow[] = [];
    for (const nodePanelName of properties?.nodes || []) {
      const panel = panelsMap[nodePanelName];
      // If we can't find the panel, just continue? Not ideal...
      if (!panel) {
        continue;
      }

      const artificialCategoryId = `node_category_${nodePanelName}`;
      const artificialPlaceholderCategoryId = `node_category_placeholder_${nodePanelName}`;

      const typedPanelData = (panel.data || {}) as NodeAndEdgeData;

      // Get a union of all the columns across all nodes
      for (const column of typedPanelData.columns || []) {
        if (columns.some((c) => c.name === column.name)) {
          continue;
        }
        columns.push(column);
      }

      // If we don't have any rows for this node type, add a placeholder
      const nodeProperties = (panel.properties || {}) as NodeProperties;
      if (!typedPanelData.rows && nodeProperties.category) {
        newProperties = set(
          newProperties || {},
          `categories["${artificialPlaceholderCategoryId}"]`,
          {
            ...nodeProperties.category,
            fold: {
              ...(nodeProperties.category.fold || {}),
              threshold: 1,
            },
          }
        );
        rows.push({
          id: nodePanelName,
          title: nodeProperties.category.title || nodeProperties.category.name,
          category: artificialPlaceholderCategoryId,
        });
        continue;
      }

      // Ensure we have category info set for each row
      for (const row of typedPanelData.rows || []) {
        // Ensure each row has an id
        if (row.id === null || row.id === undefined) {
          continue;
        }
        // Capture the ID of each row
        nodeIdLookup[row.id.toString()] = row;
        const updatedRow = row;
        if (!updatedRow.title && panel.title) {
          updatedRow.title = panel.title;
        }
        // If a row defines a category, then it is assumed to be present in the categories map
        // If there's a category defined on the node, we need to capture it
        if (!updatedRow.category) {
          if (nodeProperties.category) {
            newProperties = set(
              newProperties || {},
              `categories["${artificialCategoryId}"]`,
              nodeProperties.category
            );
          }
          updatedRow.category = artificialCategoryId;
        }
        rows.push(updatedRow);
      }
    }
    for (const edgePanelName of properties?.edges || []) {
      const panel = panelsMap[edgePanelName];
      if (!panel || !panel.data) {
        continue;
      }
      const typedPanelData = panel.data as NodeAndEdgeData;
      for (const column of typedPanelData.columns) {
        if (columns.some((c) => c.name === column.name)) {
          continue;
        }
        columns.push(column);
      }
      const artificialCategoryId = `node_category_${edgePanelName}`;
      // Ensure we have category info set for each row
      for (const row of typedPanelData.rows || []) {
        // Ensure the node this edge points to exists in the data set
        // @ts-ignore
        const from_id = has(row, "from_id") ? row.from_id.toString() : null;
        // @ts-ignore
        const to_id = has(row, "to_id") ? row.to_id.toString() : null;
        if (
          !from_id ||
          !to_id ||
          !nodeIdLookup[from_id] ||
          !nodeIdLookup[to_id]
        ) {
          continue;
        }
        const updatedRow = row;
        if (!updatedRow.title && panel.title) {
          updatedRow.title = panel.title;
        }
        // If a row defines a category, then it is assumed to be present in the categories map
        // If there's a category defined on the node, we need to capture it
        if (!updatedRow.category) {
          const edgeProperties = panel.properties as EdgeProperties;
          if (edgeProperties.category) {
            newProperties = set(
              newProperties || {},
              `categories["${artificialCategoryId}"]`,
              edgeProperties.category
            );
          }
          updatedRow.category = artificialCategoryId;
        }
        rows.push(updatedRow);
      }
    }
    return {
      data: { columns, rows },
      properties: newProperties,
    };
  }, [data, panelsMap, properties, status]);
};

export default useNodeAndEdgeData;

export { getNodeAndEdgeDataFormat };
