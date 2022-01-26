import IntegerDisplay from "../../../IntegerDisplay";
import LoadingIndicator from "../../LoadingIndicator";
import React from "react";
import { classNames } from "../../../../utils/styles";
import { getTextClasses, getWrapperClasses } from "../../Counter";
import { startCase } from "lodash";

interface ControlCounterProps {
  loading: boolean;
  status: "alarm" | "error" | "info" | "ok" | "skip";
  value: number;
}

const getCounterStyle = (status) => {
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

const getCounterLabel = (status) => {
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
      return "skip";
    default:
      return startCase(status);
  }
};

const CheckCounter = ({
  loading = true,
  status,
  value,
}: ControlCounterProps) => {
  const counterStyle = getCounterStyle(status);
  const counterLabel = getCounterLabel(status);
  const wrapperClass = getWrapperClasses(counterStyle);
  const textClass = getTextClasses(counterStyle);
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
      <dt className="text-sm font-medium truncate">{counterLabel}</dt>
      <dd className="mt-1 text-3xl font-semibold truncate">
        {loading && <LoadingIndicator />}
        {!loading && <IntegerDisplay num={value} />}
      </dd>
    </div>
  );
};

export default CheckCounter;
