import Icon from "../Icon";
import moment from "moment/moment";
import NeutralButton from "../forms/NeutralButton";
import { DashboardDataModeCLISnapshot } from "../../types";
import { saveAs } from "file-saver";
import { useDashboard } from "../../hooks/useDashboard";

const SaveSnapshotButton = () => {
  const { dashboard, dataMode, selectedDashboard, snapshot } = useDashboard();

  const saveSnapshot = () => {
    if (!dashboard || !snapshot) {
      return;
    }
    const blob = new Blob([JSON.stringify(snapshot)], {
      type: "application/json",
    });
    saveAs(blob, `${dashboard.name}.${moment().format("YYYYMMDDTHHmmss")}.sps`);
  };

  if (
    dataMode === DashboardDataModeCLISnapshot ||
    (!selectedDashboard && !snapshot)
  ) {
    return null;
  }

  return (
    <NeutralButton
      className="inline-flex items-center space-x-1"
      disabled={!dashboard || !snapshot}
      onClick={saveSnapshot}
    >
      <>
        <Icon
          className="inline-block text-foreground-lighter w-5 -mt-0.5"
          icon="camera"
        />
        <span className="hidden lg:block">Snap</span>
      </>
    </NeutralButton>
  );
};

export default SaveSnapshotButton;
