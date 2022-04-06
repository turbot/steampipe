import LoadingIndicator from "../../LoadingIndicator";
import padStart from "lodash/padStart";
import { CheckProps, CheckRunState } from "../common";
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
import { useState } from "react";

const getPadding = (depth) => {
  switch (depth) {
    case 1:
      return "pl-1";
    case 2:
      return "pl-2";
    case 3:
      return "pl-3";
    case 4:
      return "pl-4";
    case 5:
      return "pl-5";
    case 6:
      return "pl-6";
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

interface CheckNodeStatusProps {
  run_state: CheckRunState;
}

const BenchmarkIcon = ({ expanded, run_state }) => (
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

const CheckNodeStatus = ({ run_state }: CheckNodeStatusProps) => {
  return (
    <>
      {(run_state === "ready" || run_state === "started") && (
        <LoadingIndicator className="flex-shrink-0 w-5 h-5 text-foreground-lightest" />
      )}
      {run_state === "complete" && (
        <OKIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
      )}
      {run_state === "error" && (
        <ErrorIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
      )}
      {run_state === "unknown" && (
        <UnknownIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
      )}
    </>
  );
};

const ControlNode = ({ depth = 0, control }: ControlNodeProps) => {
  return (
    <div className="group flex rounded-sm">
      <div
        className={classNames(
          "px-1 space-x-2 flex flex-grow group-hover:bg-black-scale-1"
        )}
      >
        <CheckNodeStatus run_state={control.run_state} />
        <p className={classNames("", getPadding(depth))}>
          {control.title || control.name}
        </p>
      </div>
      <CheckSummary summary={control.summary} />
    </div>
  );
};

const BenchmarkNode = ({ depth = 0, benchmark }: BenchmarkNodeProps) => {
  const [expanded, setExpanded] = useState(depth < 1);
  return (
    <>
      <div className="group flex rounded-sm">
        <div
          className={classNames(
            "px-1 space-x-2 flex flex-grow group-hover:bg-black-scale-1",
            benchmark.benchmarks || benchmark.controls
              ? "cursor-pointer "
              : null
          )}
          onClick={
            benchmark.benchmarks || benchmark.controls
              ? () => setExpanded(!expanded)
              : undefined
          }
        >
          <BenchmarkIcon expanded={expanded} run_state={benchmark.run_state} />
          {/*<div className="relative flex items-center">*/}
          {/*  {!expanded && (*/}
          {/*    <ExpandBenchmarkIcon className="mr-1 inline w-5 h-5 text-foreground-lighter" />*/}
          {/*  )}*/}
          {/*  {expanded && (*/}
          {/*    <CollapseBenchmarkIcon className="mr-1 inline w-5 h-5 text-foreground-lighter" />*/}
          {/*  )}*/}
          {/*  {(benchmark.run_state === "ready" ||*/}
          {/*    benchmark.run_state === "started") && (*/}
          {/*    <LoadingIndicator className="absolute mr-1 w-5 h-5 top-px left-0 text-foreground-lightest" />*/}
          {/*  )}*/}
          {/*<LoadingIndicator className="absolute w-5 h-5 top-px left-0 text-foreground-lightest" />*/}
          {/*<CheckNodeStatus run_state={benchmark.run_state} />*/}
          <p className={classNames("", getPadding(depth))}>
            {benchmark.title || benchmark.name}
          </p>
          {/*</div>*/}
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
  const rootGroups = props.root.child_groups;
  if (!rootGroups) {
    return null;
  }
  const rootBenchmark = rootGroups[0];
  const benchmark = new BenchmarkType(
    rootBenchmark.group_id,
    rootBenchmark.title,
    rootBenchmark.description,
    rootBenchmark.child_groups,
    rootBenchmark.child_controls
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
