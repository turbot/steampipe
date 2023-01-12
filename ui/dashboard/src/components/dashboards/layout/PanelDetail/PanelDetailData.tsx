import Panel from "../Panel";
import { getComponent } from "../../index";
import { PanelDetailProps } from "./index";

const Table = getComponent("table");

const PanelDetailData = ({ definition }: PanelDetailProps) => {
  return (
    <Panel
      definition={definition}
      parentType="dashboard"
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

export default PanelDetailData;
