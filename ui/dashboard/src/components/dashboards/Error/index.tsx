import ErrorMessage from "../../ErrorMessage";

interface ErrorProps {
  error?: any;
}

const Error = ({ error }: ErrorProps) => {
  if (!error) {
    return null;
  }
  return (
    <div className="flex w-full h-full p-2 break-all bg-red-50 border-red-700 border text-red-700 justify-center items-center shadow rounded-md">
      <ErrorMessage error={error} />
    </div>
  );
};

export default Error;
