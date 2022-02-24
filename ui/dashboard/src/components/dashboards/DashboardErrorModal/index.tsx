import ErrorModal from "../../Modal/ErrorModal";
import { useDashboard } from "../../../hooks/useDashboard";

interface DashboardErrorModalProps {
  error: any;
}

const DashboardErrorModal = ({ error }: DashboardErrorModalProps) => (
  <ErrorModal error={error} title="Dashboard Error" />
);

const DashboardErrorModalWrapper = () => {
  const { error } = useDashboard();
  if (!error) {
    return null;
  }
  return <DashboardErrorModal error={error} />;
};

export default DashboardErrorModalWrapper;
