import NeutralButton from "../../../forms/NeutralButton";
import useDownloadPanelData from "../../../../hooks/useDownloadPanelData";
import { noop } from "../../../../utils/func";

const PanelDetailDataDownloadButton = ({ panelDefinition, size }) => {
  const { download, processing } = useDownloadPanelData(panelDefinition);

  return (
    <NeutralButton
      disabled={processing}
      onClick={processing ? noop : () => download()}
      size={size}
    >
      <>Download</>
    </NeutralButton>
  );
};

export default PanelDetailDataDownloadButton;
