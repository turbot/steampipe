import Children from "../common/Children";
import { PanelDetailProps } from "./index";

const PanelDetailQuery = ({ definition }: PanelDetailProps) => (
  <Children
    children={[{ ...definition, width: 12 }]}
    allowPanelExpand={false}
  />
);

export default PanelDetailQuery;
