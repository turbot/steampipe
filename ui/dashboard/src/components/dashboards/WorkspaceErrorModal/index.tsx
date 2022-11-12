import ErrorModal from "../../Modal/ErrorModal";
import { useDashboard } from "../../../hooks/useDashboard";

interface WorkspaceErrorModalProps {
  error: any;
}

const WorkspaceErrorModal = ({ error }: WorkspaceErrorModalProps) => (
  <ErrorModal error={error} title="Workspace Error" />
);

const WorkspaceErrorModalWrapper = () => {
  const { error } = useDashboard();
  if (!error) {
    return null;
  }
  return <WorkspaceErrorModal error={error} />;
};

export default WorkspaceErrorModalWrapper;
