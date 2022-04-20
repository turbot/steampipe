import { CheckSummary } from "../common";
import { classNames } from "../../../../utils/styles";

interface ProgressBarGroupProps {
  children: JSX.Element | JSX.Element[];
  className?: string;
}

interface ProgressBarProps {
  className?: string;
  value: number;
}

interface CheckSummaryChartProps {
  summary: CheckSummary;
  // rootSummary: CheckSummary;
}

const getWidth = (x, y) => {
  const percent = (x / (x + y)) * 100;
  return percent >= 0.5 ? Math.round(percent) : 1;
};

const ProgressBarGroup = ({ children, className }: ProgressBarGroupProps) => (
  <div className={classNames("flex h-3", className)}>{children}</div>
);

const ProgressBar = ({ className, value }: ProgressBarProps) => {
  if (!value) {
    return null;
  }
  return (
    <div
      className={classNames("h-3", className)}
      aria-valuenow={value}
      aria-valuemin={0}
      aria-valuemax={100}
      style={{ display: "inline-block", width: `${value}%` }}
    />
  );
};

const CheckSummaryChart = ({
  summary,
}: // rootSummary,
CheckSummaryChartProps) => {
  // const maxAlerts = rootSummary.alarm + rootSummary.error;
  // const maxNonAlerts = rootSummary.ok + rootSummary.info + rootSummary.skip;
  const { alarm, error, ok, info, skip } = summary;
  const total = alarm + error + ok + info + skip;
  // let alertsWidth = getWidth(maxAlerts, maxNonAlerts);
  // let nonAlertsWidth = getWidth(maxNonAlerts, maxAlerts);
  // if (alertsWidth > nonAlertsWidth) {
  //   alertsWidth -= 2;
  // } else {
  //   nonAlertsWidth -= 2;
  // }

  if (total === 0) {
    return null;
  }

  return (
    <div className="flex w-96">
      <ProgressBar
        className="border border-alert"
        value={(error / total) * 100}
      />
      <ProgressBar
        className="bg-alert border border-alert"
        value={(alarm / total) * 100}
      />
      <ProgressBar
        className="bg-ok border border-ok"
        value={(ok / total) * 100}
      />
      <ProgressBar
        className="bg-info border border-info"
        value={(info / total) * 100}
      />
      <ProgressBar
        className="bg-tbd border border-tbd"
        value={(skip / total) * 100}
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
