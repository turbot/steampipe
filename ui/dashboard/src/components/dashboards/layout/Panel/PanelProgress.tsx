import { classNames } from "../../../../utils/styles";
import { DashboardDataModeLive } from "../../../../types";
import { getNodeAndEdgeDataFormat } from "../../common/useNodeAndEdgeData";
import { useDashboard } from "../../../../hooks/useDashboard";
import { useMemo } from "react";
import { usePanel } from "../../../../hooks/usePanel";

const PanelProgress = ({ className }) => {
  const { definition } = usePanel();
  const { dataMode, panelsMap } = useDashboard();

  const showProgress = useMemo(
    () =>
      !definition.error &&
      (definition.panel_type === "flow" ||
        definition.panel_type === "graph" ||
        definition.panel_type === "hierarchy") &&
      getNodeAndEdgeDataFormat(definition.properties) === "NODE_AND_EDGE",
    [definition]
  );

  const progress = useMemo(() => {
    if (!showProgress) {
      return 100;
    }

    if (definition.status === "complete") {
      return 100;
    }
    const nodes: string[] = definition?.properties?.nodes || [];
    const edges: string[] = definition?.properties?.edges || [];

    if (nodes.length === 0 && edges.length === 0) {
      return 100;
    }

    const totalNodesAndEdges = nodes.length + edges.length;
    let completedNodesAndEdges = 0;
    for (const panelName of [...nodes, ...edges]) {
      const panel = panelsMap[panelName];
      if (panel && (panel.status === "complete" || panel.status === "error")) {
        completedNodesAndEdges += 1;
      }
    }

    return Math.min(
      Math.ceil((completedNodesAndEdges / totalNodesAndEdges) * 100),
      100
    );
  }, [definition, panelsMap, showProgress]);

  // We only show a progress indicator in live mode
  if (dataMode !== DashboardDataModeLive) {
    return null;
  }

  return showProgress ? (
    <div
      className={classNames(
        className,
        "w-full h-[4px] bg-dashboard-panel print:hidden"
      )}
    >
      {progress < 100 && (
        <div
          className="h-full bg-dashboard"
          style={{ width: `${progress}%` }}
        />
      )}
    </div>
  ) : null;
};

export default PanelProgress;
