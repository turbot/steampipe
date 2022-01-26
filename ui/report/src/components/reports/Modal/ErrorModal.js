import Icon from "../../Icon";
import Modal from "./index";
import { errorIcon } from "../../../constants/icons";

const ErrorModal = ({ error, title }) => {
  return (
    <Modal
      icon={
        <Icon
          className="h-6 w-6 text-red-600"
          icon={errorIcon}
          aria-hidden="true"
        />
      }
      message={error}
      title={title}
    />
  );
};

export default ErrorModal;
