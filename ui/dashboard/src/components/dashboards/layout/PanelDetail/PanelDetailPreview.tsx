import Child from "../Child";
import { PanelDetailProps } from "./index";
import { PanelDefinition } from "../../../../types";
import { useDashboard } from "../../../../hooks/useDashboard";

const PanelDetailPreview = ({
  definition: { children, name, panel_type, title, ...rest },
}: PanelDetailProps) => {
  const { panelsMap } = useDashboard();
  const layoutDefinition = { children, name, panel_type };
  const panelDefinition = {
    name,
    panel_type,
    width: 12,
    ...rest,
  } as PanelDefinition;
  return (
    <Child
      layoutDefinition={layoutDefinition}
      panelDefinition={panelDefinition}
      panelsMap={{ ...panelsMap, [name]: panelDefinition }}
      showPanelControls={false}
    />
  );
};

export default PanelDetailPreview;
