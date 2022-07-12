import IntegerDisplay from "../../../IntegerDisplay";
import LoadingIndicator from "../../LoadingIndicator";
import startCase from "lodash/startCase";
import { classNames } from "../../../../utils/styles";
import { getTextClasses, getWrapperClasses } from "../../../../utils/card";

interface ControlCardProps {
  loading: boolean;
  status: "alarm" | "error" | "info" | "ok" | "skip";
  value: number;
}

const getCardStyle = (status) => {
  switch (status) {
    case "alarm":
    case "error":
      return "alert";
    case "info":
      return "info";
    case "ok":
      return "ok";
    default:
      return "plain";
  }
};

const getCardLabel = (status) => {
  switch (status) {
    case "alarm":
      return "Alarm";
    case "error":
      return "Error";
    case "info":
      return "Info";
    case "ok":
      return "OK";
    case "skip":
      return "Skip";
    default:
      return startCase(status);
  }
};

const CheckCard = ({ loading = true, status, value }: ControlCardProps) => {
  const cardStyle = getCardStyle(status);
  const cardLabel = getCardLabel(status);
  const wrapperClass = getWrapperClasses(cardStyle);
  const textClass = getTextClasses(cardStyle);
  return (
    <div
      className={classNames(
        "col-span-1 overflow-hidden shadow rounded-lg py-2 px-4",
        wrapperClass,
        // (status === "alarm" || status === "error") && value === 0
        // value === 0 ? "opacity-60" : null,
        textClass
      )}
    >
      <dt className="text-sm font-medium truncate">{cardLabel}</dt>
      <dd className="mt-1 text-3xl font-semibold truncate">
        {loading && <LoadingIndicator />}
        {!loading && <IntegerDisplay num={value} />}
      </dd>
    </div>
  );
};

export default CheckCard;
