import Error from "../../Error";
import LoadingIndicator from "../../LoadingIndicator";
import padStart from "lodash/padStart";
import { CheckProps } from "../common";
import { classNames } from "../../../../utils/styles";
import { default as BenchmarkType } from "../common/Benchmark";
import { default as ControlType } from "../common/Control";
import {
  CollapseBenchmarkIcon,
  ErrorIcon,
  ExpandBenchmarkIcon,
  OKIcon,
  UnknownIcon,
} from "../../../../constants/icons";
import { stringToColour } from "../../../../utils/color";
import { TableView } from "../../Table";
import { useState } from "react";

const getPadding = (depth) => {
  switch (depth) {
    case 1:
      return "pl-[6px]";
    case 2:
      return "pl-[12px]";
    case 3:
      return "pl-[18px]";
    case 4:
      return "pl-[24px]";
    case 5:
      return "pl-[30px]";
    case 6:
      return "pl-[36px]";
    default:
      return "pl-0";
  }
};

interface ControlNodeProps {
  depth: number;
  control: ControlType;
}

interface BenchmarkNodeProps {
  depth: number;
  benchmark: BenchmarkType;
}

const CheckSummary = ({ summary }) => {
  return (
    <div className="flex tabular-nums space-x-4 justify-end group-hover:bg-black-scale-1 px-1">
      <pre
        className={classNames(
          "inline",
          summary.ok > 0 ? "text-ok" : "text-foreground-lightest"
        )}
      >{`${padStart(summary.ok, 5)}`}</pre>
      <pre
        className={classNames(
          "inline",
          summary.skip > 0 ? null : "text-foreground-lightest"
        )}
      >{`${padStart(summary.skip, 5)}`}</pre>
      <pre
        className={classNames(
          "inline",
          summary.info > 0 ? "text-info" : "text-foreground-lightest"
        )}
      >{`${padStart(summary.info, 5)}`}</pre>
      <pre
        className={classNames(
          "inline",
          summary.alarm > 0 ? "text-alert" : "text-foreground-lightest"
        )}
      >{`${padStart(summary.alarm, 5)}`}</pre>
      <pre
        className={classNames(
          "inline",
          summary.error > 0 ? "text-alert" : "text-foreground-lightest"
        )}
      >{`${padStart(summary.error, 5)}`}</pre>
    </div>
  );
};

const BenchmarkNodeIcon = ({ expanded, run_state }) => (
  <div className="relative flex items-center">
    {!expanded && (
      <ExpandBenchmarkIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
    )}
    {expanded && (
      <CollapseBenchmarkIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
    )}
    {(run_state === "ready" || run_state === "started") && (
      <LoadingIndicator className="absolute flex-shrink-0 w-5 h-5 top-px left-0 text-foreground-lightest" />
    )}
  </div>
);

const ControlNodeIcon = ({ expanded, results, run_state }) => {
  return expanded || results.length > 0 ? (
    <div className="relative flex items-center">
      {!expanded && (
        <ExpandBenchmarkIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
      )}
      {expanded && (
        <CollapseBenchmarkIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
      )}
      {(run_state === "ready" || run_state === "started") && (
        <LoadingIndicator className="absolute flex-shrink-0 w-5 h-5 top-px left-0 text-foreground-lightest" />
      )}
    </div>
  ) : (
    <div className="flex items-center">
      {run_state === "error" && (
        <ErrorIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
      )}
      {run_state === "complete" && (
        <OKIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
      )}
      {run_state === "unknown" && (
        <UnknownIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
      )}
      {(run_state === "ready" || run_state === "started") && (
        <LoadingIndicator className="flex-shrink-0 w-5 h-5 top-px left-0 text-foreground-lightest" />
      )}
    </div>
  );
};

const controlColumns = [
  {
    Header: "status",
    accessor: "status",
    name: "status",
    data_type_name: "CONTROL_STATUS",
    wrap: false,
  },
  {
    Header: "resource",
    accessor: "resource",
    name: "resource",
    data_type_name: "TEXT",
    wrap: false,
  },
  {
    Header: "reason",
    accessor: "reason",
    name: "reason",
    data_type_name: "TEXT",
    wrap: true,
  },
  {
    Header: "dimensions",
    accessor: "dimensions",
    name: "dimensions",
    data_type_name: "CONTROL_DIMENSIONS",
    wrap: true,
  },
];

const ControlDimension = ({ key, value }) => (
  <span
    className="rounded-md text-xs"
    style={{ color: stringToColour(value) }}
    title={`${key} = ${value}`}
  >
    {value}
  </span>
);

const ControlResults = ({ results }) => {
  return (
    <TableView
      columns={controlColumns}
      rowData={results}
      hasTopBorder={true}
      hiddenColumns={[]}
    />
  );
};

const ControlNode = ({ depth = 0, control }: ControlNodeProps) => {
  const [showControls, setShowControls] = useState(false);
  return (
    <>
      <div className="group flex rounded-sm">
        <div
          className={classNames(
            "px-1 space-x-2 flex flex-grow group-hover:bg-black-scale-1 items-center",
            showControls || control.results.length > 0 || control.run_error
              ? "cursor-pointer "
              : null,
            getPadding(depth)
          )}
          onClick={
            showControls || control.results.length > 0 || control.run_error
              ? () => setShowControls(!showControls)
              : undefined
          }
        >
          <ControlNodeIcon
            expanded={showControls}
            results={control.results}
            run_state={control.run_state}
          />
          {/*<CheckNodeStatus run_state={control.run_state} />*/}
          <p>{control.title || control.name}</p>
        </div>
        <CheckSummary summary={control.summary} />
      </div>
      {showControls && (
        <>
          {control.results.length > 0 && (
            <div className="p-4 w-full overflow-x-auto">
              <ControlResults results={control.results} />
            </div>
          )}
          {control.run_error && (
            <div className="pl-4 py-4">
              <Error error={control.run_error} />
            </div>
          )}
        </>
      )}
    </>
  );
};

const BenchmarkNode = ({ depth = 0, benchmark }: BenchmarkNodeProps) => {
  const [expanded, setExpanded] = useState(depth < 1);
  return (
    <>
      <div className="group flex rounded-sm">
        <div
          className={classNames(
            "px-1 space-x-2 flex flex-grow group-hover:bg-black-scale-1 items-center",
            benchmark.benchmarks || benchmark.controls
              ? "cursor-pointer "
              : null,
            getPadding(depth)
          )}
          onClick={
            benchmark.benchmarks || benchmark.controls
              ? () => setExpanded(!expanded)
              : undefined
          }
        >
          <BenchmarkNodeIcon
            expanded={expanded}
            run_state={benchmark.run_state}
          />
          <p>{benchmark.title || benchmark.name}</p>
        </div>
        <CheckSummary summary={benchmark.summary} />
      </div>
      {expanded && (
        <>
          {benchmark.benchmarks.map((b) => (
            <BenchmarkNode key={b.name} depth={depth + 1} benchmark={b} />
          ))}
          {benchmark.controls.map((c) => (
            <ControlNode key={c.name} depth={depth + 1} control={c} />
          ))}
        </>
      )}
    </>
  );
};

const Benchmark = (props: CheckProps) => {
  const rootGroups = props.root.groups;
  if (!rootGroups) {
    return null;
  }
  const rootBenchmark = rootGroups[0];
  const benchmark = new BenchmarkType(
    rootBenchmark.group_id,
    rootBenchmark.title,
    rootBenchmark.description,
    rootBenchmark.groups,
    rootBenchmark.controls
  );

  return (
    <div className="p-4">
      <div className="flex flex-grow"></div>
      <div className="flex text-foreground-light space-x-4 tabular-nums justify-end px-1">
        <pre className="inline">{`${padStart("OK", 5)}`}</pre>
        <pre className="inline">{`${padStart("Skip", 5)}`}</pre>
        <pre className="inline">{`${padStart("Info", 5)}`}</pre>
        <pre className="inline">{`${padStart("Alarm", 5)}`}</pre>
        <pre className="inline">{`${padStart("Error", 5)}`}</pre>
      </div>
      <BenchmarkNode depth={0} benchmark={benchmark} />
    </div>
  );
};

export default Benchmark;

export { ControlDimension };
