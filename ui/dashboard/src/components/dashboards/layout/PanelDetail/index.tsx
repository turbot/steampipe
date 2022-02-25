import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import NeutralButton from "../../../forms/NeutralButton";
import PanelQuery from "./PanelQuery";
import { PanelDefinition, useDashboard } from "../../../../hooks/useDashboard";
import Table from "../../Table";

type PanelDetailProps = {
  definition: PanelDefinition;
};

const PanelDetail = ({ definition }: PanelDetailProps) => {
  const { closePanelDetail } = useDashboard();
  return (
    <LayoutPanel definition={definition} withPadding={true}>
      <div className="col-span-11">
        <h2 className="text-2xl font-medium break-all">Panel Detail</h2>
      </div>
      <div className="col-span-1 text-right">
        <NeutralButton onClick={closePanelDetail}>
          <>
            Close<span className="ml-2 font-light text-xxs">ESC</span>
          </>
        </NeutralButton>
      </div>
      <Children
        children={[{ ...definition, width: 12 }]}
        allowPanelExpand={false}
      />
      <div className="col-span-12 grid grid-cols-12">
        <div className="col-span-12 md:col-span-6 lg:col-span-4">
          {definition.sql && <PanelQuery query={definition.sql} />}
        </div>
        <div className="col-span-12 md:col-span-6 lg:col-span-8">
          {definition.data && (
            <Table
              name={`${definition}.table.detail`}
              node_type="table"
              data={definition.data}
            />
          )}
        </div>
      </div>
    </LayoutPanel>
  );
};

export default PanelDetail;
