import { classNames } from "../../../../utils/styles";
import { DashboardDataModeLive } from "../../../../types";
import { getNodeAndEdgeDataFormat } from "../../common/useNodeAndEdgeData";
import { NodeAndEdgeProperties } from "../../common/types";
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
      getNodeAndEdgeDataFormat(
        definition.properties as NodeAndEdgeProperties
      ) === "NODE_AND_EDGE",
    [definition]
  );

  const progress = useMemo(() => {
    if (!showProgress) {
      return 100;
    }

    if (definition.status === "complete") {
      return 100;
    }

    const nodeAndEdgeProperties =
      definition?.properties as NodeAndEdgeProperties;
    const nodes: string[] = nodeAndEdgeProperties?.nodes || [];
    const edges: string[] = nodeAndEdgeProperties?.edges || [];

    if (nodes.length === 0 && edges.length === 0) {
      return 100;
    }

    const dependencyPanels = {};

    let totalThings = nodes.length + edges.length;
    let totalComplete = 0;
    let totalError = 0;
    for (const panelName of [...nodes, ...edges]) {
      const panel = panelsMap[panelName];
      if (!panel) {
        continue;
      }

      for (const dependency of panel.dependencies || []) {
        if (dependencyPanels[dependency]) {
          continue;
        }
        const dependencyPanel = panelsMap[dependency];
        if (!dependencyPanel) {
          continue;
        }
        dependencyPanels[dependency] = dependencyPanel;
        if (dependencyPanel.status === "error") {
          totalError += 1;
        } else if (dependencyPanel.status === "complete") {
          totalComplete += 1;
        }
      }

      if (panel.status === "error") {
        totalError += 1;
      } else if (panel.status === "complete") {
        totalComplete += 1;
      }
    }

    totalThings += Object.keys(dependencyPanels).length;

    return Math.min(
      Math.ceil(((totalError + totalComplete) / totalThings) * 100),
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
