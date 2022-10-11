import ErrorMessage from "../../ErrorMessage";
import { registerComponent } from "../index";
import { classNames } from "../../../utils/styles";

interface ErrorProps {
  className?: string | null;
  error?: any;
}

const Error = ({ className, error }: ErrorProps) => {
  if (!error) {
    return null;
  }
  return (
    <div
      className={classNames(
        "flex w-full h-full p-2 break-all bg-red-50 border-red-700 border text-red-700 justify-center items-center shadow rounded-md",
        className
      )}
    >
      <ErrorMessage error={error} />
    </div>
  );
};

registerComponent("error", Error);

export default Error;
