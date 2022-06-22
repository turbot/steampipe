import ErrorModal from "../../Modal/ErrorModal";
import { useDashboardNew } from "../../../hooks/refactor/useDashboard";

interface DashboardErrorModalProps {
  error: any;
}

const DashboardErrorModal = ({ error }: DashboardErrorModalProps) => (
  <ErrorModal error={error} title="Dashboard Error" />
);

const DashboardErrorModalWrapper = () => {
  const { error } = useDashboardNew();
  if (!error) {
    return null;
  }
  return <DashboardErrorModal error={error} />;
};

export default DashboardErrorModalWrapper;
