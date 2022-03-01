import Table from "../../Table";
import { PanelDetailProps } from "./index";
import { PanelProvider } from "../../../../hooks/usePanel";

const PanelDetailPreview = ({ definition }: PanelDetailProps) => {
  return (
    <PanelProvider
      definition={definition}
      allowExpand={false}
      setZoomIconClassName={() => {}}
    >
      <Table
        name={`${definition}.table.detail`}
        node_type="table"
        data={definition.data}
      />
    </PanelProvider>
  );
};

export default PanelDetailPreview;
