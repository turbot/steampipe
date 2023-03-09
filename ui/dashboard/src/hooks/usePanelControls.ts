import useDownloadPanelData from "./useDownloadPanelData";
import useSelectPanel from "./useSelectPanel";
import { IPanelControl } from "../components/dashboards/layout/Panel/PanelControls";
import { PanelDefinition } from "../types";
import { useCallback, useEffect, useState } from "react";

const usePanelControls = (definition, show = false) => {
  const { download } = useDownloadPanelData(definition as PanelDefinition);
  const { select } = useSelectPanel(definition as PanelDefinition);

  const downloadPanelData = useCallback(
    async (e) => {
      e.stopPropagation();
      await download();
    },
    [download]
  );

  const getBasePanelControls = useCallback(() => {
    const controls: IPanelControl[] = [];
    if (!show || !definition) {
      return controls;
    }
    if (definition.data) {
      controls.push({
        action: downloadPanelData,
        icon: "arrow-down-tray",
        title: "Download data",
      });
    }
    controls.push({
      action: select,
      icon: "arrows-pointing-out",
      title: "View detail",
    });
    return controls;
  }, [definition, downloadPanelData, select, show]);

  const [panelControls, setPanelControls] = useState(getBasePanelControls());

  useEffect(
    () => setPanelControls(getBasePanelControls()),
    [definition, getBasePanelControls, setPanelControls, show]
  );

  return { panelControls };
};

export default usePanelControls;
