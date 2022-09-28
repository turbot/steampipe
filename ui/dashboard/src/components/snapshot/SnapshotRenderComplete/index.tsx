import { useDashboard } from "../../../hooks/useDashboard";

const SnapshotRenderComplete = () => {
  const {
    render: { snapshotCompleteDiv },
  } = useDashboard();

  if (!snapshotCompleteDiv) {
    return null;
  }

  return <div id="snapshot-complete" className="hidden" />;
};

export default SnapshotRenderComplete;
