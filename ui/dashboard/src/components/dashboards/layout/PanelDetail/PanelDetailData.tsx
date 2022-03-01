import Table from "../../Table";
import { PanelDetailProps } from "./index";

const PanelDetailPreview = ({ definition }: PanelDetailProps) => {
  return (
    <div className="col-span-12 mt-4">
      <Table
        name={`${definition}.table.detail`}
        node_type="table"
        data={definition.data}
      />
    </div>
  );
};

export default PanelDetailPreview;
