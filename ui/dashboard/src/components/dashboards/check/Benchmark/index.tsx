import CheckGrouping from "../CheckGrouping";
import Container from "../../layout/Container";
import Error from "../../Error";
import Panel from "../../layout/Panel";
import Table from "../../Table";
import useCheckGrouping from "../../../../hooks/useCheckGrouping";
import {
  BenchmarkTreeProps,
  CheckDisplayGroup,
  CheckNode,
  CheckProps,
} from "../common";
import { default as BenchmarkType } from "../common/Benchmark";
import { LeafNodeData } from "../../common";
import { stringToColour } from "../../../../utils/color";
import { useMemo } from "react";

interface BenchmarkTableViewProps {
  benchmark: BenchmarkType;
  definition: CheckProps;
}

type InnerCheckProps = CheckProps & {
  data?: LeafNodeData;
  grouping: CheckNode;
  groupingConfig: CheckDisplayGroup[];
};

const ControlDimension = ({ dimensionKey, dimensionValue }) => (
  <span
    className="rounded-md text-xs"
    style={{ color: stringToColour(dimensionValue) }}
    title={`${dimensionKey} = ${dimensionValue}`}
  >
    {dimensionValue}
  </span>
);

const Benchmark = (props: InnerCheckProps) => {
  // const benchmarkDataTable = useMemo(() => {
  //   if (
  //     !props.benchmark ||
  //     (props.benchmark.status !== "complete" &&
  //       props.benchmark.status !== "error")
  //   ) {
  //     return undefined;
  //   }
  //   return props.benchmark.get_data_table();
  // }, [props.benchmark]);

  if (!props.grouping) {
    return null;
  }

  const summary = props.grouping.summary;

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
                  icon: "heroicons-solid:exclamation-circle",
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
                  icon: "heroicons-solid:x-circle",
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
                  grouping: props.grouping,
                  root_summary: summary,
                },
              },
            ],
          },
        ],
        // data: benchmarkDataTable,
      }}
    />
  );
};

const BenchmarkTree = (props: BenchmarkTreeProps) => {
  // const rootSummary = useMemo(() => {
  //   if (!props.properties.benchmark) {
  //     return null;
  //   }
  //   return props.properties.benchmark.summary;
  // }, [props.properties.benchmark]);

  if (!props.properties || !props.properties.root_summary) {
    return null;
  }

  return (
    <CheckGrouping
      node={props.properties.grouping}
      groupingConfig={props.properties.grouping_config}
      rootSummary={props.properties.root_summary}
    />
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
  const [grouping, groupingsConfig] = useCheckGrouping(props);

  if (!grouping) {
    return null;
  }

  console.log(grouping);

  // if (props.properties && props.properties.type === "table") {
  //   return <BenchmarkTableView benchmark={groups} definition={props} />;
  // }

  return (
    <Benchmark
      {...props}
      grouping={grouping}
      groupingConfig={groupingsConfig}
    />
  );
};

export default BenchmarkWrapper;

export { BenchmarkTree, ControlDimension };
