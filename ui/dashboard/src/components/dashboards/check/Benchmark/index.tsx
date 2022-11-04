import Card, { CardProps, CardType } from "../../Card";
import CheckGrouping from "../CheckGrouping";
import ContainerTitle from "../../titles/ContainerTitle";
import Error from "../../Error";
import Grid from "../../layout/Grid";
import Panel from "../../layout/Panel";
import { BenchmarkDefinition, PanelDefinition } from "../../../../types";
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
import { getComponent, registerComponent } from "../../index";
import { noop } from "../../../../utils/func";
import { useDashboard } from "../../../../hooks/useDashboard";
import { useMemo, useState } from "react";
import { Width } from "../../common";

const Table = getComponent("table");

interface BenchmarkTableViewProps {
  benchmark: BenchmarkType;
  definition: BenchmarkDefinition;
}

type InnerCheckProps = {
  benchmark: BenchmarkType;
  definition: BenchmarkDefinition;
  grouping: CheckNode;
  groupingConfig: CheckDisplayGroup[];
  firstChildSummaries: CheckSummary[];
  showControls: boolean;
  withTitle: boolean;
};

const Benchmark = (props: InnerCheckProps) => {
  const { dashboard, selectedDashboard } = useDashboard();
  const [showBenchmarkControls, setShowBenchmarkControls] = useState(false);
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

  const summaryCards = useMemo(() => {
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
        name: `${props.definition.name}.container.summary.ok-${totalSummary.ok}`,
        width: 2,
        display_type: totalSummary.ok > 0 ? "ok" : null,
        properties: {
          label: "OK",
          value: totalSummary.ok.toString(),
          icon: "heroicons-solid:check-circle",
        },
      },
      {
        name: `${props.definition.name}.container.summary.alarm-${totalSummary.alarm}`,
        width: 2,
        display_type: totalSummary.alarm > 0 ? "alert" : null,
        properties: {
          label: "Alarm",
          value: totalSummary.alarm.toString(),
          icon: "heroicons-solid:bell",
        },
      },
      {
        name: `${props.definition.name}.container.summary.error-${totalSummary.error}`,
        width: 2,
        display_type: totalSummary.error > 0 ? "alert" : null,
        properties: {
          label: "Error",
          value: totalSummary.error.toString(),
          icon: "heroicons-solid:exclamation-circle",
        },
      },
      {
        name: `${props.definition.name}.container.summary.info-${totalSummary.info}`,
        width: 2,
        display_type: totalSummary.info > 0 ? "info" : null,
        properties: {
          label: "Info",
          value: totalSummary.info.toString(),
          icon: "heroicons-solid:information-circle",
        },
      },
      {
        name: `${props.definition.name}.container.summary.skip-${totalSummary.skip}`,
        width: 2,
        properties: {
          label: "Skipped",
          value: totalSummary.skip.toString(),
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
        name: `${props.definition.name}.container.summary.severity-${total}`,
        width: 2,
        display_type: total > 0 ? "severity" : "",
        properties: {
          label: "Critical / High",
          value: total.toString(),
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
    <Grid
      name={props.definition.name}
      width={props.definition.width}
      events={{
        onMouseEnter: props.showControls
          ? () => setShowBenchmarkControls(true)
          : noop,
        onMouseLeave: () => setShowBenchmarkControls(false),
      }}
    >
      {!dashboard?.artificial && (
        <ContainerTitle title={props.benchmark.title} />
      )}
      <Grid name={`${props.definition.name}.container.summary`}>
        {summaryCards.map((summaryCard) => {
          const props: CardProps = {
            name: summaryCard.name,
            display_type: summaryCard.display_type as CardType,
            panel_type: "card",
            properties: summaryCard.properties,
            status: "complete",
            width: summaryCard.width as Width,
          };
          return (
            <Panel
              key={summaryCard.name}
              definition={props}
              showControls={false}
            >
              <Card {...props} />
            </Panel>
          );
        })}
      </Grid>
      <Grid name={`${props.definition.name}.container.tree`}>
        <BenchmarkTree
          name={`${props.definition.name}.container.tree.results`}
          panel_type="benchmark_tree"
          properties={{
            grouping: props.grouping,
            first_child_summaries: props.firstChildSummaries,
          }}
          status="complete"
        />
      </Grid>
    </Grid>
  );

  // return (
  //   <Container
  //     definition={{
  //       name: props.definition.name,
  //       panel_type: "container",
  //       children: [
  //         {
  //           name: `${props.definition.name}.container.summary`,
  //           panel_type: "container",
  //           children: summary_cards,
  //         },
  //         {
  //           name: `${props.definition.name}.container.tree`,
  //           panel_type: "container",
  //           children: [
  //             {
  //               name: `${props.definition.name}.container.tree.results`,
  //               panel_type: "benchmark_tree",
  //               properties: {
  //                 grouping: props.grouping,
  //                 first_child_summaries: props.firstChildSummaries,
  //               },
  //             },
  //           ],
  //         },
  //       ],
  //       data: benchmarkDataTable,
  //       title: dashboard?.artificial ? undefined : props.benchmark.title,
  //       width: props.definition.width,
  //     }}
  //     // @ts-ignore
  //     expandDefinition={{
  //       ...props.definition,
  //       title: props.benchmark.title || selectedDashboard?.title,
  //       data: benchmarkDataTable,
  //     }}
  //     showControls={true}
  //     withTitle={props.withTitle}
  //   />
  // );
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
        panel_type: "table",
        width: definition.width,
        children: definition.children,
        data: benchmarkDataTable,
        status: benchmarkDataTable ? "complete" : "ready",
      }}
      ready={!!benchmarkDataTable}
    >
      <Table
        name={`${definition.name}.table`}
        panel_type="table"
        data={benchmarkDataTable}
      />
    </Panel>
  );
};

const Inner = ({ showControls, withTitle }) => {
  const {
    benchmark,
    definition,
    grouping,
    groupingsConfig,
    firstChildSummaries,
  } = useCheckGrouping();

  if (!definition || !benchmark || !grouping) {
    return null;
  }

  if (!definition.display_type || definition.display_type === "benchmark") {
    return (
      <Benchmark
        benchmark={benchmark}
        definition={definition}
        grouping={grouping}
        groupingConfig={groupingsConfig}
        firstChildSummaries={firstChildSummaries}
        showControls={showControls}
        withTitle={withTitle}
      />
    );
    // @ts-ignore
  } else if (definition.display_type === "table") {
    return <BenchmarkTableView benchmark={benchmark} definition={definition} />;
  } else {
    return (
      <Panel
        definition={{
          name: definition.name,
          panel_type: "benchmark",
          width: definition.width,
          status: "error",
        }}
      >
        <Error
          error={`Unsupported benchmark type ${definition.display_type}`}
        />
      </Panel>
    );
  }
};

type BenchmarkProps = PanelDefinition & {
  showControls: boolean;
  withTitle: boolean;
};

const BenchmarkWrapper = (props: BenchmarkProps) => {
  return (
    <CheckGroupingProvider definition={props}>
      <Inner showControls={props.showControls} withTitle={props.withTitle} />
    </CheckGroupingProvider>
  );
};

registerComponent("benchmark", BenchmarkWrapper);

export default BenchmarkWrapper;
