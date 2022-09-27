import { useDashboard } from "../../../hooks/useDashboard";

const SnapshotRenderComplete = () => {
  const { renderSnapshotCompleteDiv } = useDashboard();

  if (!renderSnapshotCompleteDiv) {
    return null;
  }

  return <div id="snapshot-complete" />;
};

export default SnapshotRenderComplete;
