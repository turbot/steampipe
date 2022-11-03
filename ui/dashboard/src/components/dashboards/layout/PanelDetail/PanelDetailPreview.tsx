import Child from "../Child";
import { PanelDetailProps } from "./index";
import { PanelDefinition } from "../../../../types";

const PanelDetailPreview = ({
  definition: { name, panel_type, title, ...rest },
}: PanelDetailProps) => {
  const layoutDefinition = { name, panel_type };
  const panelDefinition = {
    name,
    panel_type,
    width: 12,
    ...rest,
  } as PanelDefinition;
  const panelsMap = { [name]: panelDefinition };
  return (
    <Child
      layoutDefinition={layoutDefinition}
      panelDefinition={panelDefinition}
      panelsMap={panelsMap}
      showPanelControls={false}
    />
  );
};

export default PanelDetailPreview;
