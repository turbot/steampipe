import Panel from "../Panel";
import { getComponent } from "../../index";
import { PanelDetailProps } from "./index";

const Table = getComponent("table");

const PanelDetailPreview = ({ definition }: PanelDetailProps) => {
  return (
    <Panel
      className={"overflow-x-auto"}
      definition={definition}
      showControls={false}
      forceBackground={true}
    >
      <Table
        name={`${definition}.table.detail`}
        panel_type="table"
        data={definition.data}
      />
    </Panel>
  );
};

export default PanelDetailPreview;
