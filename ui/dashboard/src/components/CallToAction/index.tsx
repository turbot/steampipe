import { classNames } from "../../utils/styles";

interface CallToActionProps {
  action?: any;
  className?: string;
  icon: any;
  size?: "lg" | "xl";
  title: JSX.Element | JSX.Element[];
  message: JSX.Element | JSX.Element[];
}

const CallToAction = ({
  action,
  className,
  icon,
  size,
  title,
  message,
}: CallToActionProps) => {
  return (
    <div
      className={classNames(
        "flex flex-col sm:flex-row p-4 sm:p-12 border border-divide rounded-lg space-x-0 sm:space-x-8 space-y-2 sm:space-y-0",
        className
      )}
    >
      <div className="flex items-center sm:block space-x-4">
        {icon}
        <h2 className="block sm:hidden font-bold leading-6">{title}</h2>
      </div>
      <div className="overflow-x-hidden w-full">
        <h2 className="hidden sm:block font-bold leading-6">{title}</h2>
        <div
          className={classNames(
            "mt-4 space-y-4 text-foreground",
            size === "xl"
              ? "max-w-full"
              : size === "lg"
              ? "max-w-2xl"
              : "max-w-lg"
          )}
        >
          {message}
        </div>
        {action && <div className="mt-6">{action}</div>}
      </div>
    </div>
  );
};

export default CallToAction;
