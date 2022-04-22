import { CheckSummary } from "../common";
import { classNames } from "../../../../utils/styles";

interface ProgressBarGroupProps {
  children: JSX.Element | JSX.Element[];
  className?: string;
}

interface ProgressBarProps {
  className?: string;
  percent: number;
}

interface CheckSummaryChartProps {
  name: string;
  summary: CheckSummary;
}

// const getWidth = (x, y) => {
//   const percent = (x / (x + y)) * 100;
//   return percent >= 0.5 ? Math.round(percent) : 1;
// };
//
// const ProgressBarGroup = ({ children, className }: ProgressBarGroupProps) => (
//   <div className={classNames("flex h-3", className)}>{children}</div>
// );

interface ValueWithIndex {
  value: number;
  percent: number;
  index: number;
}

const ensureMinPercentages = (
  name,
  values: number[] = [],
  minPercentage = 2
) => {
  // Summary here is I want to ensure each percent is >= 2% and a round number, so I'll adjust
  // all other values accordingly to ensure we total 100%
  const total = values.reduce((partial, v) => partial + v, 0);
  const valuesWithPercentAndIndex: ValueWithIndex[] = [];
  for (let i = 0; i < values.length; i++) {
    const value = values[i];
    valuesWithPercentAndIndex.push({
      value,
      percent: (value / total) * 100,
      index: i,
    });
  }
  const withMinPercentages = valuesWithPercentAndIndex.map((p) => ({
    ...p,
    percent:
      p.percent > 0 && p.percent < minPercentage ? minPercentage : p.percent,
  }));
  const flooredPercentages = withMinPercentages.map((p) => ({
    ...p,
    percent: p.percent > 0 ? Math.floor(p.percent) : p.percent,
  }));
  let diff =
    flooredPercentages.reduce((partial, v) => partial + v.percent, 0) - 100;
  const numberOfValuesToDistributeAcross = flooredPercentages.filter((p) => {
    if (diff < 0) {
      return p.percent > minPercentage && 100 - p.percent + 4 > 0;
    } else {
      return p.percent > minPercentage && p.percent - 4 > minPercentage;
    }
  }).length;
  const perItem = diff / numberOfValuesToDistributeAcross;
  // if (name === "aws_compliance.control.cis_v140_1_12") {
  //   console.log({
  //     values,
  //     total,
  //     valuesWithPercentAndIndex,
  //     withMinPercentages,
  //     flooredPercentages,
  //     numberOfValuesToDistributeAcross,
  //     perItem,
  //     diff,
  //   });
  // }
  let adjusted;
  if (diff < 0) {
    const ascending = [...flooredPercentages]
      .sort((a, b) =>
        a.percent < b.percent ? -1 : a.percent > b.percent ? 1 : 0
      )
      .map((p) => ({ ...p }));
    for (const percentageItem of ascending) {
      if (
        diff === 0 ||
        percentageItem.percent < minPercentage ||
        percentageItem.percent - 4 <= minPercentage
      ) {
        continue;
      }
      if (perItem < 0 && perItem > -1) {
        percentageItem.percent += 1;
        diff += 1;
      } else {
        percentageItem.percent -= perItem;
        diff -= perItem;
      }
    }
    adjusted = ascending
      .sort((a, b) => (a.index < b.index ? -1 : a.index > b.index ? 1 : 0))
      .map((p) => p.percent);
  } else {
    const descending = [...flooredPercentages]
      .sort((a, b) =>
        b.percent < a.percent ? -1 : b.percent > a.percent ? 1 : 0
      )
      .map((p) => ({ ...p }));
    for (const percentageItem of descending) {
      if (
        diff === 0 ||
        percentageItem.percent < minPercentage ||
        percentageItem.percent - 4 <= minPercentage
      ) {
        continue;
      }
      if (perItem > 0 && perItem < 1) {
        percentageItem.percent -= 1;
        diff -= 1;
      } else {
        percentageItem.percent -= perItem;
        diff -= perItem;
      }
    }
    adjusted = descending
      .sort((a, b) => (a.index < b.index ? -1 : a.index > b.index ? 1 : 0))
      .map((p) => p.percent);
  }
  // if (name === "aws_compliance.control.cis_v140_1_12") {
  //   console.log(adjusted);
  // }
  return adjusted;
};

const ProgressBar = ({ className, percent }: ProgressBarProps) => {
  if (!percent) {
    return null;
  }
  return (
    <div
      className={classNames("h-3", className)}
      aria-valuenow={percent}
      aria-valuemin={0}
      aria-valuemax={100}
      style={{ display: "inline-block", width: `${percent}%` }}
    />
  );
};

const CheckSummaryChart = ({ name, summary }: CheckSummaryChartProps) => {
  // const maxAlerts = rootSummary.alarm + rootSummary.error;
  // const maxNonAlerts = rootSummary.ok + rootSummary.info + rootSummary.skip;
  const [alarm, error, ok, info, skip] = ensureMinPercentages(name, [
    summary.alarm,
    summary.error,
    summary.ok,
    summary.info,
    summary.skip,
  ]);
  // let alertsWidth = getWidth(maxAlerts, maxNonAlerts);
  // let nonAlertsWidth = getWidth(maxNonAlerts, maxAlerts);
  // if (alertsWidth > nonAlertsWidth) {
  //   alertsWidth -= 2;
  // } else {
  //   nonAlertsWidth -= 2;
  // }

  if (
    summary.alarm + summary.error + summary.ok + summary.info + summary.skip ===
    0
  ) {
    return null;
  }

  return (
    <div className="flex w-96">
      <ProgressBar
        className={classNames(
          "border border-alert",
          error > 0 ? "rounded-l-sm" : null,
          skip === 0 && info === 0 && ok === 0 && alarm === 0 && error > 0
            ? "rounded-r-sm"
            : null
        )}
        percent={error}
      />
      <ProgressBar
        className={classNames(
          "bg-alert border border-alert",
          error === 0 && alarm > 0 ? "rounded-l-sm" : null,
          skip === 0 && info === 0 && ok === 0 && alarm > 0
            ? "rounded-r-sm"
            : null
        )}
        percent={alarm}
      />
      <ProgressBar
        className={classNames(
          "bg-ok border border-ok",
          error === 0 && alarm === 0 && ok > 0 ? "rounded-l-sm" : null,
          skip === 0 && info === 0 && ok > 0 ? "rounded-r-sm" : null
        )}
        percent={ok}
      />
      <ProgressBar
        className={classNames(
          "bg-info border border-info",
          error === 0 && alarm === 0 && ok === 0 && info > 0
            ? "rounded-l-sm"
            : null,
          skip === 0 && info > 0 ? "rounded-r-sm" : null
        )}
        percent={info}
      />
      <ProgressBar
        className={classNames(
          "bg-tbd border border-tbd",
          error === 0 && alarm === 0 && ok === 0 && info === 0 && error > 0
            ? "rounded-l-sm"
            : null,
          skip > 0 ? "rounded-r-sm" : null
        )}
        percent={skip}
      />
      {/*<div className="my-auto px-0" style={{ width: `${alertsWidth}%` }}>*/}
      {/*  <ProgressBarGroup className="flex-row-reverse">*/}
      {/*    <ProgressBar className="bg-alert border border-alert" value={alarm} />*/}
      {/*    <ProgressBar className="border border-alert" value={error} />*/}
      {/*  </ProgressBarGroup>*/}
      {/*</div>*/}
      {/*<div className="h-6 w-0 border-l border-black-scale-4" />*/}
      {/*<div className="my-auto px-0" style={{ width: `${nonAlertsWidth}%` }}>*/}
      {/*  <ProgressBarGroup>*/}
      {/*    <ProgressBar className="bg-ok border border-ok" value={ok} />*/}
      {/*    <ProgressBar className="bg-info border border-info" value={info} />*/}
      {/*    <ProgressBar className="bg-tbd border border-tbd" value={skip} />*/}
      {/*  </ProgressBarGroup>*/}
      {/*</div>*/}
    </div>
  );
};

export default CheckSummaryChart;
