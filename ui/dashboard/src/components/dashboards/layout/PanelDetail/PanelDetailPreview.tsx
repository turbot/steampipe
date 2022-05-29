import Children from "../common/Children";
import { PanelDetailProps } from "./index";

const PanelDetailPreview = ({ definition }: PanelDetailProps) => (
  <Children
    children={[{ ...definition, width: 12 }]}
    allowPanelExpand={false}
    isExpanded={true}
    withTitle={false}
  />
);

export default PanelDetailPreview;
