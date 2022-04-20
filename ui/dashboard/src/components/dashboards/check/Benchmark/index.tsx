import CheckSummaryChart from "../CheckSummaryChart";
import Container from "../../layout/Container";
import Error from "../../Error";
import Panel from "../../layout/Panel";
import padEnd from "lodash/padEnd";
import Table from "../../Table";
import { BenchmarkTreeProps, CheckProps, CheckSummary } from "../common";
import { classNames } from "../../../../utils/styles";
import { default as BenchmarkType } from "../common/Benchmark";
import { default as ControlType } from "../common/Control";
import { get } from "lodash";
import { LeafNodeData } from "../../common";
import { stringToColour } from "../../../../utils/color";
import { useMemo, useState } from "react";

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
  rootSummary: CheckSummary;
}

interface BenchmarkNodeProps {
  depth: number;
  benchmark: BenchmarkType;
  rootSummary: CheckSummary;
}

interface BenchmarkTableViewProps {
  benchmark: BenchmarkType;
  definition: CheckProps;
}

// const CheckSummary = ({ summary }) => {
//   return (
//     <div className="flex tabular-nums space-x-4 justify-end group-hover:bg-black-scale-1 px-1">
//       <pre
//         className={classNames(
//           "inline",
//           summary.ok > 0 ? "text-ok" : "text-foreground-lightest"
//         )}
//       >{`${padStart(summary.ok, 5)}`}</pre>
//       <pre
//         className={classNames(
//           "inline",
//           summary.alarm > 0 ? "text-alert" : "text-foreground-lightest"
//         )}
//       >{`${padStart(summary.alarm, 5)}`}</pre>
//       <pre
//         className={classNames(
//           "inline",
//           summary.error > 0 ? "text-alert" : "text-foreground-lightest"
//         )}
//       >{`${padStart(summary.error, 5)}`}</pre>
//       <pre
//         className={classNames(
//           "inline",
//           summary.info > 0 ? "text-info" : "text-foreground-lightest"
//         )}
//       >{`${padStart(summary.info, 5)}`}</pre>
//       <pre
//         className={classNames(
//           "inline",
//           summary.skip > 0 ? null : "text-foreground-lightest"
//         )}
//       >{`${padStart(summary.skip, 5)}`}</pre>
//     </div>
//   );
// };

// const BenchmarkNodeIcon = ({ expanded, run_state }) => (
//   <div className="relative flex items-center">
//     {!expanded && (
//       <ExpandBenchmarkIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
//     )}
//     {expanded && (
//       <CollapseBenchmarkIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
//     )}
//     {(run_state === "ready" || run_state === "started") && (
//       <LoadingIndicator className="absolute flex-shrink-0 w-5 h-5 top-px left-0 text-foreground-lightest" />
//     )}
//   </div>
// );
//
// const ControlNodeIcon = ({ expanded, run_state }) => {
//   return (
//     <div className="relative flex items-center">
//       {!expanded && (
//         <ExpandBenchmarkIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
//       )}
//       {expanded && (
//         <CollapseBenchmarkIcon className="flex-shrink-0 w-5 h-5 text-foreground-lighter" />
//       )}
//       {(run_state === "ready" || run_state === "started") && (
//         <LoadingIndicator className="absolute flex-shrink-0 w-5 h-5 top-px left-0 text-foreground-lightest" />
//       )}
//     </div>
//   );
// };

// const controlColumns = [
//   {
//     Header: "status",
//     accessor: "status",
//     name: "status",
//     data_type_name: "CONTROL_STATUS",
//     wrap: false,
//   },
//   {
//     Header: "resource",
//     accessor: "resource",
//     name: "resource",
//     data_type_name: "TEXT",
//     wrap: false,
//   },
//   {
//     Header: "reason",
//     accessor: "reason",
//     name: "reason",
//     data_type_name: "TEXT",
//     wrap: true,
//   },
//   {
//     Header: "dimensions",
//     accessor: "dimensions",
//     name: "dimensions",
//     data_type_name: "CONTROL_DIMENSIONS",
//     wrap: true,
//   },
// ];

const ControlDimension = ({ dimensionKey, dimensionValue }) => (
  <span
    className="rounded-md text-xs"
    style={{ color: stringToColour(dimensionValue) }}
    title={`${dimensionKey} = ${dimensionValue}`}
  >
    {dimensionValue}
  </span>
);

const ControlResultStatus = ({ status }) => {
  let textClass;
  switch (status) {
    case "alarm":
    case "error":
      textClass = "text-alert";
      break;
    case "ok":
      textClass = "text-ok";
      break;
    case "info":
      textClass = "text-info";
      break;
    default:
      textClass = "text-tbd";
      break;
  }
  return (
    <pre>
      <span className={classNames(textClass, "uppercase")}>
        {padEnd(status, 5)}
      </span>
      <span className="text-foreground-lightest">:</span>
    </pre>
  );
};

const ControlResult = ({ depth, result }) => {
  return (
    <div className="group flex rounded-sm">
      <ControlIndentMarker depth={depth} />
      <div className="pr-1 space-x-2 flex flex-grow group-hover:bg-black-scale-1 items-baseline">
        <div className="flex truncate">
          <ControlResultStatus status={result.status} />
          <span className="px-1">{result.reason}</span>
        </div>
        <div className="flex-grow border-b border-dotted" />
        <div className="space-x-2">
          {(result.dimensions || []).map((dimension) => (
            <ControlDimension
              key={dimension.key}
              dimensionKey={dimension.key}
              dimensionValue={dimension.value}
            />
          ))}
        </div>
      </div>
    </div>
  );
};

const ControlError = ({ depth, error }) => {
  return (
    <div className="group flex rounded-sm">
      <ControlIndentMarker depth={depth} />
      <div className="pr-1 space-x-2 flex flex-grow group-hover:bg-black-scale-1 items-baseline">
        <div className="flex truncate">
          <ControlResultStatus status="error" />
          <span className="px-1 text-alert">{error}</span>
        </div>
      </div>
    </div>
  );
};

const ControlResults = ({ depth, results }) => {
  return results.map((result) => (
    <ControlResult key={result.resource} depth={depth} result={result} />
  ));
};

const ControlNode = ({ depth = 0, control, rootSummary }: ControlNodeProps) => {
  const [showControls, setShowControls] = useState(false);
  return (
    <>
      <div className="group flex rounded-sm py-2">
        <NodeIndentMarker depth={depth} expanded={showControls} />
        <div
          className={classNames(
            "pr-1 space-x-2 flex flex-grow flex-nowrap truncate group-hover:bg-black-scale-1 items-center",
            showControls || control.results.length > 0 || control.run_error
              ? "cursor-pointer "
              : null
          )}
          onClick={
            showControls || control.results.length > 0 || control.run_error
              ? () => setShowControls(!showControls)
              : undefined
          }
        >
          <p>{control.title || control.name}</p>
        </div>
        <CheckSummaryChart
          summary={control.summary}
          rootSummary={rootSummary}
        />
        {/*<CheckSummary summary={control.summary} />*/}
      </div>
      {showControls && (
        <>
          {control.results.length > 0 && (
            <>
              <ControlIndentMarker depth={depth + 1} />
              <ControlResults depth={depth} results={control.results} />
              <ControlIndentMarker depth={depth} />
            </>
          )}
          {control.run_error && (
            <>
              <ControlIndentMarker depth={depth + 1} />
              <ControlError depth={depth + 1} error={control.run_error} />
            </>
            // <div className="pl-4 py-4">
            //   <Error error={control.run_error} />
            // </div>
          )}
        </>
      )}
    </>
  );
};

const ControlIndentMarker = ({ depth }) => {
  const barArray = Array(depth > 0 ? depth : 0).fill("|");
  const joined = barArray.join(" ");

  return (
    <span className="flex-shrink-0 font-mono text-foreground-lightest">
      {joined}
      {depth > 0 ? <span>&nbsp;</span> : ""}
    </span>
  );
};

const NodeIndentMarker = ({ depth, expanded }) => {
  const barArray = Array(depth > 0 ? depth - 1 : 0).fill("|");
  // const plusArray = Array(depth > 0 ? 1 : 0).fill(expanded ? "-" : "+");
  const plusArray = Array(depth > 0 ? 1 : 0).fill("+");
  const combinedArray = [...barArray, ...plusArray];
  const joined = combinedArray.join(" ");

  return (
    <span className="flex-shrink-0 font-mono text-foreground-lightest">
      {joined}
      {depth > 0 ? <span>&nbsp;</span> : ""}
    </span>
  );
};

const BenchmarkNode = ({
  depth = 0,
  benchmark,
  rootSummary,
}: BenchmarkNodeProps) => {
  const [expanded, setExpanded] = useState(depth < 1);

  return (
    <>
      <div className="group flex rounded-sm py-2">
        <NodeIndentMarker depth={depth} expanded={expanded} />
        <div
          className={classNames(
            "pr-1 space-x-2 flex flex-grow group-hover:bg-black-scale-1 items-center",
            benchmark.benchmarks || benchmark.controls
              ? "cursor-pointer "
              : null
            // getPadding(depth)
          )}
          onClick={
            benchmark.benchmarks || benchmark.controls
              ? () => setExpanded(!expanded)
              : undefined
          }
        >
          {/*<BenchmarkNodeIcon*/}
          {/*  expanded={expanded}*/}
          {/*  run_state={benchmark.run_state}*/}
          {/*/>*/}
          <p>{benchmark.title || benchmark.name}</p>
        </div>
        <CheckSummaryChart
          summary={benchmark.summary}
          rootSummary={rootSummary}
        />
        {/*<CheckSummary summary={benchmark.summary} />*/}
      </div>
      {expanded && (
        <>
          {benchmark.benchmarks.map((b) => (
            <BenchmarkNode
              key={b.name}
              depth={depth + 1}
              benchmark={b}
              rootSummary={rootSummary}
            />
          ))}
          {benchmark.controls.map((c) => (
            <ControlNode
              key={c.name}
              depth={depth + 1}
              control={c}
              rootSummary={rootSummary}
            />
          ))}
        </>
      )}
    </>
  );
};

type InnerCheckProps = CheckProps & {
  data?: LeafNodeData;
  benchmark: BenchmarkType | null;
};

const Benchmark = (props: InnerCheckProps) => {
  const benchmarkDataTable = useMemo(() => {
    if (
      !props.benchmark ||
      (props.benchmark.run_state !== "complete" &&
        props.benchmark.run_state !== "error")
    ) {
      return undefined;
    }
    return props.benchmark.get_data_table();
  }, [props.benchmark]);

  if (!props.benchmark) {
    return null;
  }

  const summary = props.benchmark.summary;

  return (
    <Container
      allowChildPanelExpand={false}
      allowExpand={true}
      definition={{
        name: props.name,
        node_type: "container",
        children: [
          {
            name: `${props.name}.container.summary`,
            node_type: "container",
            allow_child_panel_expand: false,
            children: [
              {
                node_type: "card",
                name: `${props.name}.container.summary.ok-${summary.ok}`,
                width: 2,
                properties: {
                  label: "OK",
                  value: summary.ok,
                  type: "ok",
                },
              },
              {
                node_type: "card",
                name: `${props.name}.container.summary.alarm-${summary.alarm}`,
                width: 2,
                properties: {
                  label: "Alarm",
                  value: summary.alarm,
                  type: summary.alarm > 0 ? "alert" : null,
                  icon: "heroicons-solid:x-circle",
                },
              },
              {
                node_type: "card",
                name: `${props.name}.container.summary.error-${summary.error}`,
                width: 2,
                properties: {
                  label: "Error",
                  value: summary.error,
                  type: summary.error > 0 ? "alert" : null,
                  icon: "heroicons-solid:exclamation-circle",
                },
              },
              {
                node_type: "card",
                name: `${props.name}.container.summary.info-${summary.info}`,
                width: 2,
                properties: {
                  label: "Info",
                  value: summary.info,
                  type: "info",
                },
              },
              {
                node_type: "card",
                name: `${props.name}.container.summary.skip-${summary.skip}`,
                width: 2,
                properties: {
                  label: "Skipped",
                  value: summary.skip,
                  icon: "heroicons-solid:arrow-circle-right",
                },
              },
            ],
          },
          {
            name: `${props.name}.container.tree`,
            node_type: "container",
            allow_child_panel_expand: false,
            children: [
              {
                name: `${props.name}.container.tree.results`,
                node_type: "benchmark_tree",
                properties: {
                  benchmark: props.benchmark,
                },
              },
            ],
          },
        ],
        data: benchmarkDataTable,
      }}
    />
  );
};

const BenchmarkTree = (props: BenchmarkTreeProps) => {
  const rootSummary = useMemo(() => {
    if (!props.properties.benchmark) {
      return null;
    }
    return props.properties.benchmark.summary;
  }, [props.properties.benchmark]);

  if (!rootSummary) {
    return null;
  }

  return (
    <div className="p-4">
      {/*// <div className="flex flex-grow"></div>*/}
      {/*//{" "}*/}
      {/*<div className="flex text-foreground-light space-x-4 tabular-nums justify-end px-1">*/}
      {/*  // <pre className="inline">{`${padStart("OK", 5)}`}</pre>*/}
      {/*  // <pre className="inline">{`${padStart("Alarm", 5)}`}</pre>*/}
      {/*  // <pre className="inline">{`${padStart("Error", 5)}`}</pre>*/}
      {/*  // <pre className="inline">{`${padStart("Info", 5)}`}</pre>*/}
      {/*  // <pre className="inline">{`${padStart("Skip", 5)}`}</pre>*/}
      {/*  //{" "}*/}
      {/*</div>*/}
      <BenchmarkNode
        depth={0}
        benchmark={props.properties.benchmark}
        rootSummary={rootSummary}
      />
    </div>
  );
};

const BenchmarkTableView = ({
  benchmark,
  definition,
}: BenchmarkTableViewProps) => {
  const benchmarkDataTable = useMemo(
    () => benchmark.get_data_table(),
    [benchmark]
  );

  return (
    <Panel
      definition={{
        name: definition.name,
        node_type: "table",
        width: definition.width,
        data: benchmarkDataTable,
      }}
      ready={!!benchmarkDataTable}
    >
      <Table
        name={`${definition.name}.table`}
        node_type="table"
        data={benchmarkDataTable}
      />
    </Panel>
  );
};

const BenchmarkWrapper = (props: CheckProps) => {
  const rootBenchmark = get(props, "execution_tree.root.groups[0]", null);

  const benchmark = useMemo(() => {
    if (!rootBenchmark) {
      return null;
    }

    return new BenchmarkType(
      rootBenchmark.group_id,
      rootBenchmark.title,
      rootBenchmark.description,
      rootBenchmark.groups,
      rootBenchmark.controls
    );
  }, [rootBenchmark]);

  if (!benchmark) {
    return null;
  }

  if (rootBenchmark.type === "table") {
    return <BenchmarkTableView benchmark={benchmark} definition={props} />;
  }

  return <Benchmark {...props} benchmark={benchmark} />;
};

export default BenchmarkWrapper;

export { BenchmarkTree, ControlDimension };

// const res = {
//   action: "leaf_node_complete",
//   dashboard_node: {
//     name: "mike.chart.dashboard_benchmarks_anonymous_chart_0",
//     sql: "      select region, count(*) as total from aws_s3_bucket group by region order by total asc\n",
//     data: {
//       columns: [
//         {
//           name: "region",
//           data_type_name: "TEXT",
//         },
//         {
//           name: "total",
//           data_type_name: "INT8",
//         },
//       ],
//       rows: [
//         ["us-west-2", 1],
//         ["ap-northeast-2", 1],
//         ["ap-south-1", 1],
//         ["ap-southeast-1", 1],
//         ["ap-southeast-2", 1],
//         ["ca-central-1", 1],
//         ["us-west-1", 1],
//         ["ap-northeast-1", 1],
//         ["eu-north-1", 1],
//         ["eu-west-1", 1],
//         ["eu-west-2", 1],
//         ["eu-west-3", 1],
//         ["sa-east-1", 1],
//         ["us-east-2", 1],
//         ["eu-central-1", 2],
//         ["us-east-1", 15],
//       ],
//     },
//     properties: {
//       type: "column",
//     },
//     node_type: "chart",
//     dashboard: "mike.dashboard.benchmarks",
//     source_definition:
//       '  chart {\n    type = "column"\n    sql = <<EOQ\n      select region, count(*) as total from aws_s3_bucket group by region order by total asc\n    EOQ\n  }',
//   },
//   execution_id: "0x14001696e10",
// };
