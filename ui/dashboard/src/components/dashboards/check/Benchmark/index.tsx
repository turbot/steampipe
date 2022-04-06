import padStart from "lodash/padStart";
import { CheckProps } from "../common";
import { default as BenchmarkType } from "../common/Benchmark";
import { default as ControlType } from "../common/Control";
import { useState } from "react";
import { classNames } from "../../../../utils/styles";

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
    <div className="col-span-8 space-x-4 tabular-nums">
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

const ControlNode = ({ depth = 0, control }: ControlNodeProps) => {
  return (
    <>
      <div
        className={classNames(
          getPadding(depth),
          "col-span-4 font-medium truncate"
        )}
      >
        {control.title || control.name}
      </div>
      <CheckSummary summary={control.summary} />
    </>
  );
};

const BenchmarkNode = ({ depth = 0, benchmark }: BenchmarkNodeProps) => {
  const [expanded, setExpanded] = useState(depth < 1);
  return (
    <>
      <div
        className={classNames(
          getPadding(depth),
          "col-span-4 font-medium truncate",
          benchmark.benchmarks || benchmark.controls ? "cursor-pointer " : null
        )}
        onClick={
          benchmark.benchmarks || benchmark.controls
            ? () => setExpanded(!expanded)
            : undefined
        }
      >
        {benchmark.title || benchmark.name}
      </div>
      <CheckSummary summary={benchmark.summary} />
      {expanded && (
        <>
          {benchmark.benchmarks.map((b) => (
            <BenchmarkNode key={b.name} depth={depth + 1} benchmark={b} />
          ))}
          {benchmark.controls.map((c) => (
            <div className="col-span-12 grid grid-cols-12 gap-x-4">
              <ControlNode key={c.name} depth={depth + 1} control={c} />
            </div>
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
    rootBenchmark.group_ip,
    rootBenchmark.title,
    rootBenchmark.description,
    rootBenchmark.child_groups,
    rootBenchmark.child_controls
  );

  return (
    <div className="p-4 grid grid-cols-12 gap-x-4 gap-y-1">
      <div className="col-span-4"></div>
      <div className="col-span-8 text-foreground-light space-x-4 tabular-nums">
        <pre className="inline">{`${padStart("OK", 5)}`}</pre>
        <pre className="inline">{`${padStart("Skip", 5)}`}</pre>
        <pre className="inline">{`${padStart("Info", 5)}`}</pre>
        <pre className="inline">{`${padStart("Alarm", 5)}`}</pre>
        <pre className="inline">{`${padStart("Error", 5)}`}</pre>
      </div>
      <BenchmarkNode depth={0} benchmark={benchmark} />
    </div>
  );
  // return <>{JSON.stringify(benchmark.summary)}</>;
  // const summary =
  //   props.execution_tree?.progress?.control_row_status_summary ||
  //   ({} as CheckLeafNodeDataGroupSummaryStatus);
  // const loading = !summary;
  // // const { loading, summary } = useMemo(() => {
  // //   const summary = props.execution_tree?.progress?.control_row_status_summary;
  // //   if (!summary) {
  // //     return {
  // //       loading: true,
  // //       summary: {} as CheckLeafNodeDataGroupSummaryStatus,
  // //     };
  // //   }
  // //   return { loading: false, summary };
  // // }, [props.execution_tree]);
  //
  // return (
  //   <LayoutPanel
  //     definition={{
  //       name: props.name,
  //       width: props.width,
  //     }}
  //   >
  //     <div className="col-span-12 grid grid-cols-5 gap-4">
  //       <CheckCard loading={loading} status="ok" value={summary.ok} />
  //       <CheckCard loading={loading} status="skip" value={summary.skip} />
  //       <CheckCard loading={loading} status="info" value={summary.info} />
  //       <CheckCard loading={loading} status="alarm" value={summary.alarm} />
  //       <CheckCard loading={loading} status="error" value={summary.error} />
  //     </div>
  //   </LayoutPanel>
  // );
};

export default Benchmark;
