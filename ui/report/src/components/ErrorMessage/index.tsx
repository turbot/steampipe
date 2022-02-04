const getErrorMessage = (error: any, fallbackMessage: string) => {
  if (!error) {
    return fallbackMessage;
  }
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

interface ErrorMessageProps {
  error?: any;
  fallbackMessage?: string;
}

const ErrorMessage = ({
  error,
  fallbackMessage = "An unknown error occurred.",
}: ErrorMessageProps) => {
  return <>{getErrorMessage(error, fallbackMessage)}</>;
};

export default ErrorMessage;
