import isNil from "lodash/isNil";
import isObject from "lodash/isObject";

interface IntegerDisplayProps {
  num: number | null;
  className?: string;
  startAt?: "k" | "100k" | "m";
  withTitle?: boolean;
}

// Adapted from https://stackoverflow.com/questions/42658221/vanilla-javascript-function-to-round-to-nearest-hundred-or-100k-plus-symbol
// const truncate = (num: number) => {
//   const p = Math.floor(Math.log(num) / Math.LN10);
//   const l = Math.min(2, Math.ceil(p / 3));
//   return (
//     Math.pow(10, p - l * 3) * +(num / Math.pow(10, p)).toFixed(1) +
//     ["", "k", "M"][l]
//   );
// };

const ten_million = 10000000;
const one_million = 1000000;
const one_hundred_thousand = 100000;
const one_thousand = 1000;

const roundToThousands = (num) => {
  return (num / 1000).toFixed(1).replace(/\.0$/, "") + "k";
};

const roundToNearestThousand = (num) => {
  const toNearest1k = Math.round(num / one_thousand) * one_thousand;
  if (toNearest1k >= one_million) {
    return roundToMillions(num);
  }
  return roundToThousands(toNearest1k);
};

const roundToMillions = (num, fixedDigits = 2) => {
  return (num / one_million).toFixed(fixedDigits).replace(/\.0+$/, "") + "M";
};

const IntegerDisplay = ({
  num,
  className,
  startAt = "m",
  withTitle = true,
}: IntegerDisplayProps) => {
  const levels = {
    k: one_thousand,
    "100k": one_hundred_thousand,
    m: one_million,
  };
  const numberFormatter = (num) => {
    if (isNil(num) || isObject(num)) {
      return;
    }

    const threshold = levels[startAt];

    if (threshold <= one_million && num >= ten_million) {
      return roundToMillions(num, 1);
    }
    if (threshold <= one_million && num >= one_million) {
      return roundToMillions(num);
    }
    if (threshold <= one_hundred_thousand && num >= one_hundred_thousand) {
      return roundToNearestThousand(num);
    }
    if (threshold <= one_thousand) {
      if (num >= one_million) {
        return roundToMillions(num);
      }
      if (num >= one_hundred_thousand) {
        return roundToNearestThousand(num);
      }
      if (num >= one_thousand) {
        return roundToThousands(num);
      }
    }
    // display number as-is but separated with commas
    return num.toLocaleString();
  };
  return (
    <span
      className={className}
      title={withTitle && num ? num.toLocaleString() : undefined}
    >
      {numberFormatter(num)}
    </span>
  );
};

export default IntegerDisplay;
