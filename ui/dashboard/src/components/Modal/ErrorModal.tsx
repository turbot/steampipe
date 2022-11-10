import ErrorMessage from "../ErrorMessage";
import Modal from "./index";
import { ErrorIcon } from "../../constants/icons";

const ErrorModal = ({ error, title }) => (
  <Modal
    icon={<ErrorIcon className="h-8 w-8 text-red-600" aria-hidden="true" />}
    message={
      <div className="break-all">
        <ErrorMessage error={error} />
      </div>
    }
    title={title}
  />
);

export default ErrorModal;
