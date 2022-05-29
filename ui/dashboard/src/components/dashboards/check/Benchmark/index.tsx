import CheckGrouping from "../CheckGrouping";
import Container from "../../layout/Container";
import Error from "../../Error";
import Panel from "../../layout/Panel";
import Table from "../../Table";
import {
  BenchmarkTreeProps,
  CheckDisplayGroup,
  CheckNode,
  CheckSummary,
} from "../common";
import {
  CheckGroupingProvider,
  useCheckGrouping,
} from "../../../../hooks/useCheckGrouping";
import { default as BenchmarkType } from "../common/Benchmark";
import { PanelDefinition, useDashboard } from "../../../../hooks/useDashboard";
import { stringToColour } from "../../../../utils/color";
import { useMemo } from "react";

interface BenchmarkTableViewProps {
  benchmark: BenchmarkType;
  definition: PanelDefinition;
}

type InnerCheckProps = {
  benchmark: BenchmarkType;
  definition: PanelDefinition;
  grouping: CheckNode;
  groupingConfig: CheckDisplayGroup[];
  firstChildSummaries: CheckSummary[];
  withTitle: boolean;
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
  const { selectedDashboard } = useDashboard();
  const benchmarkDataTable = useMemo(() => {
    if (
      !props.benchmark ||
      !props.grouping ||
      props.grouping.status !== "complete"
    ) {
      return undefined;
    }
    return props.benchmark.get_data_table();
  }, [props.benchmark, props.grouping]);

  const summary_cards = useMemo(() => {
    if (!props.grouping) {
      return [];
    }

    const totalSummary = props.firstChildSummaries.reduce(
      (cumulative, current) => {
        cumulative.error += current.error;
        cumulative.alarm += current.alarm;
        cumulative.ok += current.ok;
        cumulative.info += current.info;
        cumulative.skip += current.skip;
        return cumulative;
      },
      { error: 0, alarm: 0, ok: 0, info: 0, skip: 0 }
    );
    const summary_cards = [
      {
        node_type: "card",
        name: `${props.definition.name}.container.summary.ok-${totalSummary.ok}`,
        width: 2,
        properties: {
          label: "OK",
          value: totalSummary.ok,
          type: totalSummary.ok > 0 ? "ok" : null,
          icon: "heroicons-solid:check-circle",
        },
      },
      {
        node_type: "card",
        name: `${props.definition.name}.container.summary.alarm-${totalSummary.alarm}`,
        width: 2,
        properties: {
          label: "Alarm",
          value: totalSummary.alarm,
          type: totalSummary.alarm > 0 ? "alert" : null,
          icon: "heroicons-solid:bell",
        },
      },
      {
        node_type: "card",
        name: `${props.definition.name}.container.summary.error-${totalSummary.error}`,
        width: 2,
        properties: {
          label: "Error",
          value: totalSummary.error,
          type: totalSummary.error > 0 ? "alert" : null,
          icon: "heroicons-solid:exclamation-circle",
        },
      },
      {
        node_type: "card",
        name: `${props.definition.name}.container.summary.info-${totalSummary.info}`,
        width: 2,
        properties: {
          label: "Info",
          value: totalSummary.info,
          type: totalSummary.info > 0 ? "info" : null,
          icon: "heroicons-solid:information-circle",
        },
      },
      {
        node_type: "card",
        name: `${props.definition.name}.container.summary.skip-${totalSummary.skip}`,
        width: 2,
        properties: {
          label: "Skipped",
          value: totalSummary.skip,
          icon: "heroicons-solid:arrow-circle-right",
        },
      },
    ];

    const severity_summary = props.grouping.severity_summary;
    const criticalRaw = severity_summary["critical"];
    const highRaw = severity_summary["high"];
    const critical = criticalRaw || 0;
    const high = highRaw || 0;

    // If we have at least 1 critical or undefined control defined in this run
    if (criticalRaw !== undefined || highRaw !== undefined) {
      const total = critical + high;
      summary_cards.push({
        node_type: "card",
        name: `${props.definition.name}.container.summary.severity-${total}`,
        width: 2,
        properties: {
          label: "Critical / High",
          value: total,
          type: total > 0 ? "severity" : "",
          icon: "heroicons-solid:exclamation",
        },
      });
    }
    return summary_cards;
  }, [props.firstChildSummaries, props.grouping, props.definition.name]);

  if (!props.grouping) {
    return null;
  }

  return (
    <Container
      allowChildPanelExpand={false}
      allowExpand={true}
      definition={{
        name: props.definition.name,
        node_type: "container",
        children: [
          {
            name: `${props.definition.name}.container.summary`,
            node_type: "container",
            allow_child_panel_expand: false,
            children: summary_cards,
          },
          {
            name: `${props.definition.name}.container.tree`,
            node_type: "container",
            allow_child_panel_expand: false,
            children: [
              {
                name: `${props.definition.name}.container.tree.results`,
                node_type: "benchmark_tree",
                properties: {
                  grouping: props.grouping,
                  first_child_summaries: props.firstChildSummaries,
                },
              },
            ],
          },
        ],
        data: benchmarkDataTable,
        title: props.definition.title,
        width: props.definition.width,
      }}
      expandDefinition={{
        ...props.definition,
        title: props.definition.title || selectedDashboard?.title,
        data: benchmarkDataTable,
      }}
      withTitle={props.withTitle}
    />
  );
};

const BenchmarkTree = (props: BenchmarkTreeProps) => {
  if (!props.properties || !props.properties.first_child_summaries) {
    return null;
  }

  return <CheckGrouping node={props.properties.grouping} />;
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

const Inner = ({ withTitle }) => {
  const {
    benchmark,
    definition,
    grouping,
    groupingsConfig,
    firstChildSummaries,
    rootBenchmark,
  } = useCheckGrouping();

  if (!benchmark || !grouping || !rootBenchmark) {
    return null;
  }

  if (!rootBenchmark.type || rootBenchmark.type === "benchmark") {
    return (
      <Benchmark
        benchmark={benchmark}
        definition={definition}
        grouping={grouping}
        groupingConfig={groupingsConfig}
        firstChildSummaries={firstChildSummaries}
        withTitle={withTitle}
      />
    );
  } else if (rootBenchmark.type === "table") {
    return <BenchmarkTableView benchmark={benchmark} definition={definition} />;
  } else {
    return (
      <Panel
        definition={{
          name: definition.name,
          node_type: "benchmark",
          width: definition.width,
        }}
      >
        <Error error={`Unsupported benchmark type ${rootBenchmark.type}`} />
      </Panel>
    );
  }
};

type BenchmarkProps = PanelDefinition & {
  withTitle: boolean;
};

const BenchmarkWrapper = (props: BenchmarkProps) => {
  return (
    <CheckGroupingProvider definition={props}>
      <Inner withTitle={props.withTitle} />
    </CheckGroupingProvider>
  );
};

export default BenchmarkWrapper;

export { BenchmarkTree, ControlDimension };
