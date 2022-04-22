import CheckGrouping from "../CheckGrouping";
import Container from "../../layout/Container";
import Error from "../../Error";
import Panel from "../../layout/Panel";
import Table from "../../Table";
import { BenchmarkTreeProps, CheckProps } from "../common";
import { default as BenchmarkType } from "../common/Benchmark";
import { get } from "lodash";
import { LeafNodeData } from "../../common";
import { stringToColour } from "../../../../utils/color";
import { useMemo } from "react";

// interface ControlNodeProps {
//   depth: number;
//   control: ControlType;
// }
//
// interface BenchmarkNodeProps {
//   depth: number;
//   benchmark: BenchmarkType;
// }

interface BenchmarkTableViewProps {
  benchmark: BenchmarkType;
  definition: CheckProps;
}

// const getPadding = (depth) => {
//   switch (depth) {
//     case 1:
//       return "pl-[6px]";
//     case 2:
//       return "pl-[12px]";
//     case 3:
//       return "pl-[18px]";
//     case 4:
//       return "pl-[24px]";
//     case 5:
//       return "pl-[30px]";
//     case 6:
//       return "pl-[36px]";
//     default:
//       return "pl-0";
//   }
// };

// const getMargin = (depth) => {
//   switch (depth) {
//     case 1:
//       return "ml-[6px]";
//     case 2:
//       return "ml-[12px]";
//     case 3:
//       return "ml-[18px]";
//     case 4:
//       return "ml-[24px]";
//     case 5:
//       return "ml-[30px]";
//     case 6:
//       return "ml-[36px]";
//     default:
//       return "ml-0";
//   }
// };

const ControlDimension = ({ dimensionKey, dimensionValue }) => (
  <span
    className="rounded-md text-xs"
    style={{ color: stringToColour(dimensionValue) }}
    title={`${dimensionKey} = ${dimensionValue}`}
  >
    {dimensionValue}
  </span>
);

// const ControlResultStatus = ({ status }) => {
//   let textClass;
//   switch (status) {
//     case "alarm":
//     case "error":
//       textClass = "text-alert";
//       break;
//     case "ok":
//       textClass = "text-ok";
//       break;
//     case "info":
//       textClass = "text-info";
//       break;
//     default:
//       textClass = "text-tbd";
//       break;
//   }
//   return (
//     <pre>
//       <span className={classNames(textClass, "uppercase")}>
//         {padEnd(status, 5)}
//       </span>
//       <span className="text-foreground-lightest">:</span>
//     </pre>
//   );
// };

// const ControlResult = ({ depth, result }) => {
//   return (
//     <div className="group flex rounded-sm">
//       <ControlIndentMarker depth={depth} />
//       <div className="pr-1 space-x-2 flex flex-grow group-hover:bg-black-scale-1 items-baseline">
//         <div className="flex truncate">
//           <ControlResultStatus status={result.status} />
//           <span className="px-1">{result.reason}</span>
//         </div>
//         <div className="flex-grow border-b border-dotted" />
//         <div className="space-x-2">
//           {(result.dimensions || []).map((dimension) => (
//             <ControlDimension
//               key={dimension.key}
//               dimensionKey={dimension.key}
//               dimensionValue={dimension.value}
//             />
//           ))}
//         </div>
//       </div>
//     </div>
//   );
// };

// const ControlError = ({ depth, error }) => {
//   return (
//     <div className="group flex rounded-sm">
//       <ControlIndentMarker depth={depth} />
//       <div className="pr-1 space-x-2 flex flex-grow group-hover:bg-black-scale-1 items-baseline">
//         <div className="flex truncate">
//           <ControlResultStatus status="error" />
//           <span className="px-1 text-alert">{error}</span>
//         </div>
//       </div>
//     </div>
//   );
// };

// const ControlResults = ({ depth, results }) => {
//   return results.map((result) => (
//     <ControlResult key={result.resource} depth={depth} result={result} />
//   ));
// };

// const ControlNode = ({ depth = 0, control }: ControlNodeProps) => {
//   const [showControls, setShowControls] = useState(false);
//   return (
//     <>
//       <div className="group flex rounded-sm py-2 items-center">
//         {/*<NodeIndentMarker depth={depth} />*/}
//         <div
//           className={classNames(
//             "pr-1 space-x-2 flex flex-grow flex-nowrap truncate group-hover:bg-black-scale-1 items-center",
//             showControls || control.results.length > 0 || control.run_error
//               ? "cursor-pointer "
//               : null
//             // getPadding(depth)
//           )}
//           onClick={
//             showControls || control.results.length > 0 || control.run_error
//               ? () => setShowControls(!showControls)
//               : undefined
//           }
//         >
//           <p>{control.title || control.name}</p>
//         </div>
//         <CheckSummaryChart name={control.name} summary={control.summary} />
//       </div>
//       {showControls && (
//         <>
//           {control.results.length > 0 && (
//             <>
//               <ControlIndentMarker depth={depth + 1} />
//               <ControlResults depth={depth} results={control.results} />
//               <ControlIndentMarker depth={depth} />
//             </>
//           )}
//           {control.run_error && (
//             <>
//               <ControlIndentMarker depth={depth + 1} />
//               <ControlError depth={depth + 1} error={control.run_error} />
//             </>
//           )}
//         </>
//       )}
//     </>
//   );
// };

// const ControlIndentMarker = ({ depth }) => {
//   const barArray = Array(depth > 0 ? depth : 0).fill("|");
//   const joined = barArray.join(" ");
//
//   return (
//     <span className="flex-shrink-0 font-mono text-foreground-lightest">
//       {joined}
//       {depth > 0 ? <span>&nbsp;</span> : ""}
//     </span>
//   );
// };

// const NodeIndentMarker = ({ depth }) => {
//   // return (
//   //   <div className="flex-shrink-0 border-l border-black-scale-3 h-3 w-1" />
//   // );
//
//   const barArray = Array(depth > 0 ? depth - 1 : 0).fill("|");
//   const plusArray = Array(depth > 0 ? 1 : 0).fill("+");
//   const combinedArray = [...barArray, ...plusArray];
//   const joined = combinedArray.join(" ");
//
//   return (
//     <span className="flex-shrink-0 font-mono text-foreground-lightest">
//       {joined}
//       {depth > 0 ? <span>&nbsp;</span> : ""}
//     </span>
//   );
// };

// const BenchmarkNode = ({ depth = 0, benchmark }: BenchmarkNodeProps) => {
//   const [expanded, setExpanded] = useState(depth < 1);
//
//   return (
//     <>
//       <div className="group flex rounded-sm items-center">
//         {/*<NodeIndentMarker depth={depth} />*/}
//         <div
//           className={classNames(
//             "pr-1 space-x-2 flex flex-grow group-hover:bg-black-scale-1 items-center",
//             benchmark.benchmarks || benchmark.controls
//               ? "cursor-pointer "
//               : null
//           )}
//           onClick={
//             benchmark.benchmarks || benchmark.controls
//               ? () => setExpanded(!expanded)
//               : undefined
//           }
//         >
//           <p>{benchmark.title || benchmark.name}</p>
//         </div>
//         <CheckSummaryChart name={benchmark.name} summary={benchmark.summary} />
//       </div>
//       {expanded && (
//         <div className={classNames("border-l", getPadding(depth))}>
//           {benchmark.benchmarks.map((b) => (
//             <BenchmarkNode key={b.name} depth={depth + 1} benchmark={b} />
//           ))}
//           {benchmark.controls.map((c) => (
//             <ControlNode key={c.name} depth={depth + 1} control={c} />
//           ))}
//         </div>
//       )}
//     </>
//   );
// };

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
    <CheckGrouping
      node={props.properties.benchmark}
      rootSummary={rootSummary}
    />
  );

  // return (
  //   <div className="p-4">
  //     <BenchmarkNode depth={0} benchmark={props.properties.benchmark} />
  //   </div>
  // );
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
      0,
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
