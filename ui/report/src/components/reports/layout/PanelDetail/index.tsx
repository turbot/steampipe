import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import NeutralButton from "../../../forms/NeutralButton";
import { PanelDefinition, useReport } from "../../../../hooks/useReport";

type PanelDetailProps = {
  definition: PanelDefinition;
};

const PanelDetail = ({ definition }: PanelDetailProps) => {
  const { closePanelDetail } = useReport();
  return (
    <LayoutPanel definition={definition} withPadding={true}>
      <div className="col-span-11">
        <h2 className="text-2xl font-medium break-all">Panel Detail</h2>
      </div>
      <div className="col-span-1">
        <NeutralButton onClick={closePanelDetail}>
          <>
            Close<span className="ml-2 font-light text-xxs">ESC</span>
          </>
        </NeutralButton>
      </div>
      <Children children={[definition]} showPanelExpand={false} />
    </LayoutPanel>
  );
};

export default PanelDetail;
