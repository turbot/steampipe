import Children from "../Children";
import { PanelDetailProps } from "./index";

const PanelDetailPreview = ({ definition }: PanelDetailProps) => (
  <Children
    children={[{ ...definition, width: 12 }]}
    showPanelControls={false}
  />
);

export default PanelDetailPreview;
