import { DashboardRunState, useDashboard } from "../../../hooks/useDashboard";
import { LeafNodeData, LeafNodeDataColumn, LeafNodeDataRow } from "./index";
import { NodeAndEdgeProperties } from "./types";
import { useMemo } from "react";

const useNodeAndEdgeData = (
  data: LeafNodeData | undefined,
  properties: NodeAndEdgeProperties | undefined,
  status: DashboardRunState
) => {
  const { panelsMap } = useDashboard();
  return useMemo(() => {
    let memoData: LeafNodeData | null = null;
    if (status === "complete" && data) {
      memoData = data;
    } else if (status !== "complete") {
      const columns: LeafNodeDataColumn[] = [];
      const rows: LeafNodeDataRow[] = [];
      const nodePanelNames = (properties?.nodes || []).map((n) => n.name);
      const edgePanelNames = (properties?.edges || []).map((n) => n.name);
      for (const nodePanelName of nodePanelNames) {
        const panel = panelsMap[nodePanelName];
        if (!panel || !panel.data) {
          continue;
        }
        for (const column of panel.data.columns) {
          if (columns.some((c) => c.name === column.name)) {
            continue;
          }
          columns.push(column);
        }
        rows.push(...(panel.data.rows || []));
      }
      for (const edgePanelName of edgePanelNames) {
        const panel = panelsMap[edgePanelName];
        if (!panel || !panel.data) {
          continue;
        }
        for (const column of panel.data.columns) {
          if (columns.some((c) => c.name === column.name)) {
            continue;
          }
          columns.push(column);
        }
        rows.push(...(panel.data.rows || []));
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
