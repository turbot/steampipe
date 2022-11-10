import has from "lodash/has";
import set from "lodash/set";
import { DashboardRunState } from "../../../types";
import { EdgeProperties, NodeAndEdgeProperties, NodeProperties } from "./types";
import {
  NodeAndEdgeData,
  NodeAndEdgeDataColumn,
  NodeAndEdgeDataFormat,
  NodeAndEdgeDataRow,
  NodeAndEdgeStatus,
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

const panelStateToCategoryState = (status: DashboardRunState) => {
  return status === "error"
    ? "error"
    : status === "complete"
    ? "complete"
    : "pending";
};

const useNodeAndEdgeData = (
  data: NodeAndEdgeData | undefined,
  properties: NodeAndEdgeProperties | undefined,
  status: DashboardRunState
) => {
  const { panelsMap } = useDashboard();
  return useMemo(() => {
    const dataFormat = getNodeAndEdgeDataFormat(properties);
    if (dataFormat === "LEGACY") {
      if (status === "complete") {
        return data ? { data, dataFormat, properties } : null;
      }
      return null;
    }

    const nodeAndEdgeStatus: NodeAndEdgeStatus = {
      categories: {},
      nodes: [],
      edges: [],
    };

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

      if (nodeProperties.category && nodeProperties.category.name) {
        if (!nodeAndEdgeStatus.categories[nodeProperties.category.name]) {
          nodeAndEdgeStatus.categories[nodeProperties.category.name] = {
            id: nodeProperties.category.name,
            title: nodeProperties.category.title,
            state: panelStateToCategoryState(panel.status || "ready"),
          };
        } else if (panel.status !== "complete") {
          nodeAndEdgeStatus.categories[nodeProperties.category.name].state =
            panelStateToCategoryState(panel.status || "ready");
        }
      }

      // Ensure we have category info set for each row
      const nodeDataRows = typedPanelData.rows || [];

      nodeAndEdgeStatus.nodes.push({
        id: nodePanelName,
        title: nodePanelName.split(".").pop(),
        state: panelStateToCategoryState(panel.status || "ready"),
        count: nodeDataRows.length,
      });

      for (const row of nodeDataRows) {
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

      const edgeProperties = (panel.properties || {}) as EdgeProperties;
      if (edgeProperties.category) {
        // @ts-ignore
        if (!nodeAndEdgeStatus.categories[edgeProperties.category.name]) {
          // @ts-ignore
          nodeAndEdgeStatus.categories[edgeProperties.category.name] = {
            id: edgeProperties.category.name,
            title: edgeProperties.category.title,
            state:
              panel.status === "error"
                ? "error"
                : panel.status === "complete"
                ? "complete"
                : "pending",
          };
        } else if (panel.status !== "complete") {
          // @ts-ignore
          nodeAndEdgeStatus.categories[edgeProperties.category.name].state =
            panel.status === "error" ? "error" : "pending";
        }
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
      dataFormat,
      properties: newProperties,
      status: nodeAndEdgeStatus,
    };
  }, [data, panelsMap, properties, status]);
};

export default useNodeAndEdgeData;

export { getNodeAndEdgeDataFormat };
