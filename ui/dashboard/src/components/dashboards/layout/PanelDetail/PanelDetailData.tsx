import Panel from "../Panel";
import Table from "../../Table";
import { PanelDetailProps } from "./index";

const PanelDetailPreview = ({ definition }: PanelDetailProps) => {
  return (
    <Panel
      definition={definition}
      allowExpand={false}
      forceBackground={true}
      withOverflow={true}
      withTitle={false}
    >
      <Table
        name={`${definition}.table.detail`}
        node_type="table"
        data={definition.data}
      />
    </Panel>
  );
};

export default PanelDetailPreview;
