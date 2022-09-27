import { useDashboard } from "../../../hooks/useDashboard";

const SnapshotRenderComplete = () => {
  const {
    render: { showSnapshotCompleteDiv },
  } = useDashboard();

  if (!showSnapshotCompleteDiv) {
    return null;
  }

  return <div id="snapshot-complete" className="hidden" />;
};

export default SnapshotRenderComplete;
