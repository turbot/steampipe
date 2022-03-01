import Table from "../../Table";
import { PanelDetailProps } from "./index";

const PanelDetailPreview = ({ definition }: PanelDetailProps) => {
  return (
    <Table
      name={`${definition}.table.detail`}
      node_type="table"
      data={definition.data}
    />
  );
};

export default PanelDetailPreview;
