const getErrorMessage = (error: any, fallbackMessage: string) => {
  if (typeof error === "string") {
    return error;
  }
  if (error.message) {
    return error.message;
  }
  if (error.Message) {
    return error.Message;
  }
  return fallbackMessage;
};

interface ErrorProps {
  error?: any;
  fallbackMessage?: string;
}

const Error = ({
  error,
  fallbackMessage = "An unknown error occurred.",
}: ErrorProps) => {
  if (!error) {
    return null;
  }
  return (
    <div className="flex w-full h-full p-2 break-all bg-red-50 border-red-700 border text-red-700 justify-center items-center shadow rounded-md">
      {getErrorMessage(error, fallbackMessage)}
    </div>
  );
};

export default Error;
