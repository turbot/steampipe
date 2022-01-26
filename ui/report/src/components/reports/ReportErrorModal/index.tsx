import ErrorModal from "../Modal/ErrorModal";
import { useReport } from "../../../hooks/useReport";

interface ReportErrorModalProps {
  error: any;
}

const ReportErrorModal = ({ error }: ReportErrorModalProps) => (
  <ErrorModal error={error} title="Workspace Error" />
);

const ReportErrorModalWrapper = () => {
  const { error } = useReport();
  if (!error) {
    return null;
  }
  return <ReportErrorModal error={error} />;
};

export default ReportErrorModalWrapper;
