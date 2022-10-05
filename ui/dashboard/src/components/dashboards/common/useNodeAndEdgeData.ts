import { DashboardRunState, useDashboard } from "../../../hooks/useDashboard";
import {
  NodeAndEdgeData,
  NodeAndEdgeDataColumn,
  NodeAndEdgeDataFormat,
  NodeAndEdgeDataRow,
} from "../graphs/types";
import { NodeAndEdgeProperties } from "./types";
import { useMemo } from "react";

// Categories may be sourced from a node, an edge, a flow, a graph or a hierarchy
// A node or edge can define exactly 1 category, which covers all rows of data that don't define a category
// Any node/edge/flow/graph/hierarchy data row can define a category - in that event, the category must come from the category map for the composing resource

const useNodeAndEdgeData = (
  dataFormat: NodeAndEdgeDataFormat,
  data: NodeAndEdgeData | undefined,
  properties: NodeAndEdgeProperties | undefined,
  status: DashboardRunState
): NodeAndEdgeData | null => {
  const { panelsMap } = useDashboard();
  const res = useMemo(() => {
    if (dataFormat === "LEGACY") {
      if (status === "complete") {
        return data ? data : null;
      }
      return null;
    }

    const columns: NodeAndEdgeDataColumn[] = [];
    const rows: NodeAndEdgeDataRow[] = [];
    const nodePanelNames = (properties?.nodes || []).map((n) => n.name);
    const edgePanelNames = (properties?.edges || []).map((n) => n.name);
    for (const nodePanelName of nodePanelNames) {
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
      rows.push(...(typedPanelData.rows || []));
    }
    for (const edgePanelName of edgePanelNames) {
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
      rows.push(...(typedPanelData.rows || []));
    }
    return {
      columns,
      rows,
    };
  }, [dataFormat, data, panelsMap, properties, status]);

  console.log({ dataFormat, res, panelsMap, data, properties, status });

  return res;
};

export default useNodeAndEdgeData;
