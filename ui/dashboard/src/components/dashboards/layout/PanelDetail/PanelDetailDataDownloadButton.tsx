import NeutralButton from "../../../forms/NeutralButton";
import { noop } from "lodash";
import { useState } from "react";

const PanelDetailDataDownloadButton = ({ downloadQueryData, size }) => {
  const [downloading, setDownloading] = useState(false);

  const downloadData = async () => {
    setDownloading(true);
    downloadQueryData();
    setDownloading(false);
  };

  return (
    <NeutralButton
      disabled={downloading}
      onClick={downloading ? noop : () => downloadData()}
      size={size}
    >
      <>Download Data</>
    </NeutralButton>
  );
};

export default PanelDetailDataDownloadButton;
