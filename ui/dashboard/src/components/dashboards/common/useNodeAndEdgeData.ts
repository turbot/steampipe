import { DashboardRunState, useDashboard } from "../../../hooks/useDashboard";
import {
  NodeAndEdgeData,
  NodeAndEdgeDataColumn,
  NodeAndEdgeDataRow,
} from "../graphs/types";
import { NodeAndEdgeProperties } from "./types";
import { useMemo } from "react";

// Categories may be sourced from a node, an edge, a flow, a graph or a hierarchy
// A node or edge can define exactly 1 category, which covers all rows of data that don't define a category
// Any node/edge/flow/graph/hierarchy data row can define a category - in that event, the category must come from the category map for the composing resource

const useNodeAndEdgeData = (
  data: NodeAndEdgeData | undefined,
  properties: NodeAndEdgeProperties | undefined,
  status: DashboardRunState
): NodeAndEdgeData | null => {
  const { panelsMap } = useDashboard();
  return useMemo(() => {
    console.log({ panelsMap, data, properties, status });
    let memoData: NodeAndEdgeData | null = null;
    if (status === "complete" && data) {
      memoData = data;
    } else if (status !== "complete") {
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
      memoData = {
        columns,
        rows,
      };
    }
    return memoData;
  }, [panelsMap, data, properties, status]);
};

export default useNodeAndEdgeData;
