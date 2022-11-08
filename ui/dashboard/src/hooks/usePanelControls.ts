import { PanelControl } from "../components/dashboards/layout/Panel/PanelControls";
import { useCallback, useEffect, useState } from "react";
import useSelectPanel from "./useSelectPanel";
import { PanelDefinition } from "../types";
import useDownloadPanelData from "./useDownloadPanelData";

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

  const getBasePanelControls = () => {
    const controls: PanelControl[] = [];
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
  };

  const [panelControls, setPanelControls] = useState(getBasePanelControls());

  useEffect(() => setPanelControls(getBasePanelControls()), [definition]);

  return { panelControls };
};

export default usePanelControls;
