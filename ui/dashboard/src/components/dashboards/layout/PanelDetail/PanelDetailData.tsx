import Panel from "../Panel";
import { getComponent } from "../../index";
import { PanelDetailProps } from "./index";
const Table = getComponent("table");

const PanelDetailPreview = ({ definition }: PanelDetailProps) => {
  return (
    <Panel
      layoutDefinition={definition}
      allowExpand={false}
      forceBackground={true}
      withOverflow={true}
      withTitle={false}
    >
      {() => (
        <Table
          name={`${definition}.table.detail`}
          panel_type="table"
          data={definition.data}
        />
      )}
    </Panel>
  );
};

export default PanelDetailPreview;
