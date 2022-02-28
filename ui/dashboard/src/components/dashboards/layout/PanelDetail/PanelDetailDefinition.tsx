import Children from "../common/Children";
import { PanelDetailProps } from "./index";

const PanelDetailDefinition = ({ definition }: PanelDetailProps) => (
  <Children
    children={[{ ...definition, width: 12 }]}
    allowPanelExpand={false}
  />
);

export default PanelDetailDefinition;
