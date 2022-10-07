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

const getDataFormat = (
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
  const { dataMode, panelsMap } = useDashboard();
  return useMemo(() => {
    if (getDataFormat(properties) === "LEGACY") {
      if (status === "complete") {
        return data ? { data, properties } : null;
      }
      return null;
    }

    let newProperties = properties;
    const columns: NodeAndEdgeDataColumn[] = [];
    const rows: NodeAndEdgeDataRow[] = [];
    for (const nodePanelName of properties?.nodes || []) {
      const panel = panelsMap[nodePanelName];
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
      const artificialCategoryId = `node_category_${nodePanelName}`;
      // Ensure we have category info set for each row
      for (const row of typedPanelData.rows || []) {
        const updatedRow = row;
        if (!updatedRow.title && panel.title) {
          updatedRow.title = panel.title;
        }
        // If a row defines a category, then it is assumed to be present in the categories map
        // If there's a category defined on the node, we need to capture it
        if (!updatedRow.category) {
          const nodeProperties = panel.properties as NodeProperties;
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
