import { isNil, isObject } from "lodash";

interface IntegerDisplayProps {
  num: number | null;
  className?: string;
  startAt?: "k" | "m";
}

const IntegerDisplay = ({
  num,
  className,
  startAt = "m",
}: IntegerDisplayProps) => {
  const levels = { k: 1000, m: 1000000 };
  const numberFormatter = (num) => {
    if (isNil(num) || isObject(num)) {
      return;
    }
    if (levels[startAt] <= 1000000 && num >= 1000000) {
      return (num / 1000000).toFixed(1).replace(/\.0$/, "") + "M";
    }
    if (levels[startAt] <= 1000) {
      if (num >= 1000000) {
        return Math.round(num / 1000) + "k";
      }
      if (num >= 1000) {
        return (num / 1000).toFixed(1).replace(/\.0$/, "") + "k";
      }
    }
    // display number as is but separated with commas
    return num.toLocaleString();
  };
  return <span className={className}>{numberFormatter(num)}</span>;
};

export default IntegerDisplay;
