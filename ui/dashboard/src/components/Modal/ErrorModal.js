import ErrorMessage from "../ErrorMessage";
import Modal from "./index";
import { ErrorIcon } from "../../constants/icons";

const ErrorModal = ({ error, title }) => {
  return (
    <Modal
      icon={<ErrorIcon className="h-8 w-8 text-red-600" aria-hidden="true" />}
      message={<ErrorMessage error={error} />}
      title={title}
    />
  );
};

export default ErrorModal;
